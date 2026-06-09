//go:build integration

package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/utils"
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
	lookingGlassDiscErr    error
)

// discoverLookingGlassTarget lists MCRs on staging and returns the first one
// whose Looking Glass reports at least one BGP session in the UP state.
//
// It distinguishes "no target available" (nil target, nil error — skip) from a
// real environment problem (non-nil error — fail). A 404/not-found from the
// Looking Glass endpoint means it isn't available for that MCR and the scan
// continues; any other error (auth, 5xx) aborts and surfaces so CI fails rather
// than silently skipping coverage. The result is cached per process so the
// (slow) discovery scan runs once.
func discoverLookingGlassTarget(t *testing.T, client *megaport.Client) (*lookingGlassTarget, error) {
	t.Helper()
	lookingGlassTargetOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		mcrs, err := client.MCRService.ListMCRs(ctx, &megaport.ListMCRsRequest{})
		if err != nil {
			lookingGlassDiscErr = fmt.Errorf("ListMCRs failed during looking-glass discovery: %w", err)
			return
		}

		for _, m := range mcrs {
			if m == nil {
				continue
			}
			// Skip decommissioned/cancelled MCRs, as the list commands do. Looking
			// Glass calls against an inactive MCR can return non-404 errors that
			// would otherwise abort discovery.
			switch m.ProvisioningStatus {
			case megaport.STATUS_DECOMMISSIONED, megaport.STATUS_CANCELLED, utils.StatusDecommissioning:
				continue
			}
			sessions, err := client.MCRLookingGlassService.ListBGPSessions(ctx, m.UID)
			if err != nil {
				if megaport.IsServiceNotFoundError(err) {
					// Looking Glass not available for this MCR; keep scanning.
					continue
				}
				// Auth/5xx/etc — a real failure, not "no target". Abort so the
				// test fails instead of skipping with a misleading message.
				lookingGlassDiscErr = fmt.Errorf("ListBGPSessions failed for MCR %s: %w", m.UID, err)
				return
			}
			for _, s := range sessions {
				if s == nil {
					continue
				}
				// Require a usable SessionID — the bgp-neighbor-routes test needs
				// it, so a target without one is no better than no target.
				if s.Status == megaport.BGPSessionStatusUp && s.SessionID != "" {
					lookingGlassTargetVal = &lookingGlassTarget{
						mcrUID:    m.UID,
						sessionID: s.SessionID,
					}
					return
				}
			}
		}
	})
	return lookingGlassTargetVal, lookingGlassDiscErr
}

// requireLookingGlassTarget returns a BGP-enabled MCR, fails the test on a real
// discovery error, or skips when staging simply has no BGP-enabled MCR.
func requireLookingGlassTarget(t *testing.T, client *megaport.Client) *lookingGlassTarget {
	t.Helper()
	target, err := discoverLookingGlassTarget(t, client)
	require.NoError(t, err, "looking-glass target discovery failed")
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
