//go:build integration

package mcr

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lookingGlassTarget holds an MCR with at least one active (UP) BGP session,
// plus that session's ID for the bgp-neighbor-routes subcommand.
type lookingGlassTarget struct {
	mcrUID    string
	sessionID string
}

var (
	lookingGlassTargetOnce sync.Once
	lookingGlassTargetVal  *lookingGlassTarget
)

// discoverLookingGlassTarget lists MCRs on staging and returns the first one
// whose Looking Glass reports at least one BGP session in the UP state.
// Returns nil when no such MCR exists (e.g. the Looking Glass endpoint is not
// available on staging, in which case every MCR returns 404). The result is
// cached per process so the (slow) discovery scan runs once.
func discoverLookingGlassTarget(t *testing.T, client *megaport.Client) *lookingGlassTarget {
	t.Helper()
	lookingGlassTargetOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		mcrs, err := client.MCRService.ListMCRs(ctx, &megaport.ListMCRsRequest{})
		if err != nil {
			t.Logf("looking-glass discovery: ListMCRs failed: %v", err)
			return
		}

		for _, m := range mcrs {
			sessions, err := client.MCRLookingGlassService.ListBGPSessions(ctx, m.UID)
			if err != nil {
				// The Looking Glass endpoint may not exist for this MCR
				// (404) or be transiently unavailable; skip and keep scanning.
				continue
			}
			for _, s := range sessions {
				if s.Status == megaport.BGPSessionStatusUp {
					lookingGlassTargetVal = &lookingGlassTarget{
						mcrUID:    m.UID,
						sessionID: s.SessionID,
					}
					return
				}
			}
		}
	})
	return lookingGlassTargetVal
}

// requireLookingGlassTarget returns a BGP-enabled MCR or skips the test with a
// clear message when none is available on staging.
func requireLookingGlassTarget(t *testing.T, client *megaport.Client) *lookingGlassTarget {
	t.Helper()
	target := discoverLookingGlassTarget(t, client)
	if target == nil {
		t.Skip("no BGP-enabled MCR available on staging (no MCR reports an active Looking Glass BGP session); skipping Looking Glass integration test")
	}
	return target
}

func integrationLookingGlassCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "looking-glass"}
	cmd.Flags().String("protocol", "", "")
	cmd.Flags().String("ip", "", "")
	return cmd
}

func TestIntegration_LookingGlassBGPSessions(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	origFmt := output.GetOutputFormat()
	defer output.SetOutputFormat(origFmt)

	target := requireLookingGlassTarget(t, client)

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLookingGlassBGPSessions(integrationLookingGlassCmd(), []string{target.mcrUID}, true, "json")
	})
	require.NoError(t, err, "list BGP sessions output: %s", captured)

	var sessions []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &sessions), "output should be valid JSON: %s", captured)
	require.NotEmpty(t, sessions, "discovered MCR should report at least one BGP session")

	for _, s := range sessions {
		assert.Contains(t, s, "session_id")
		assert.Contains(t, s, "neighbor_address")
		assert.Contains(t, s, "status")
	}
}

func TestIntegration_LookingGlassIPRoutes(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	origFmt := output.GetOutputFormat()
	defer output.SetOutputFormat(origFmt)

	target := requireLookingGlassTarget(t, client)

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLookingGlassIPRoutes(integrationLookingGlassCmd(), []string{target.mcrUID}, true, "json")
	})
	require.NoError(t, err, "list IP routes output: %s", captured)

	var routes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &routes), "output should be valid JSON: %s", captured)

	// An MCR with active BGP sessions may still have an empty routing table at
	// the moment of the call; assert shape only when routes are present.
	for _, r := range routes {
		assert.Contains(t, r, "prefix")
		assert.Contains(t, r, "next_hop")
		assert.Contains(t, r, "protocol")
	}
}

func TestIntegration_LookingGlassBGPRoutes(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	origFmt := output.GetOutputFormat()
	defer output.SetOutputFormat(origFmt)

	target := requireLookingGlassTarget(t, client)

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLookingGlassBGPRoutes(integrationLookingGlassCmd(), []string{target.mcrUID}, true, "json")
	})
	require.NoError(t, err, "list BGP routes output: %s", captured)

	var routes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &routes), "output should be valid JSON: %s", captured)

	for _, r := range routes {
		assert.Contains(t, r, "prefix")
		assert.Contains(t, r, "next_hop")
		assert.Contains(t, r, "as_path")
	}
}

func TestIntegration_LookingGlassBGPNeighborRoutes(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	origFmt := output.GetOutputFormat()
	defer output.SetOutputFormat(origFmt)

	target := requireLookingGlassTarget(t, client)

	// bgp-neighbor-routes takes three positional args: MCR UID, session ID, and
	// direction ("advertised" or "received"). Use the session ID discovered from
	// the BGP-sessions endpoint. "received" is the safe choice — an UP session
	// is receiving the neighbor's routes even if the MCR advertises nothing.
	args := []string{target.mcrUID, target.sessionID, "received"}

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLookingGlassBGPNeighborRoutes(integrationLookingGlassCmd(), args, true, "json")
	})
	require.NoError(t, err, "list BGP neighbor routes output: %s", captured)

	var routes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &routes), "output should be valid JSON: %s", captured)

	for _, r := range routes {
		assert.Contains(t, r, "prefix")
		assert.Contains(t, r, "next_hop")
		assert.Contains(t, r, "as_path")
	}
}
