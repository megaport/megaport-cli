//go:build integration

package apply

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationLocationID is the staging data center used for all apply lifecycle
// tests. Location 67 matches the canonical example used across the CLI test
// suite and supports the 1G port speed these tests exercise.
const integrationLocationID = 67

// 90s allows for apply's sequential teardown: VXCs are deleted and polled to
// DECOMMISSIONING before the ports they sit on can be removed.
const (
	cleanupStatusTimeout  = 90 * time.Second
	cleanupStatusInterval = 2 * time.Second
)

func init() {
	// Integration runs poll real staging status; 2s keeps lifecycle tests
	// responsive without the 10s production default. Only compiled under the
	// integration build tag, so unit tests keep the default.
	provisionPollInterval = 2 * time.Second
}

func generateUniqueID(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 4)
	_, err := crypto_rand.Read(buf)
	require.NoError(t, err, "failed to read crypto/rand entropy")
	return hex.EncodeToString(buf)
}

// applyIntegrationCmd builds a cobra.Command carrying the flags ApplyConfig reads.
func applyIntegrationCmd(t *testing.T, file string, dryRun, yes, rollback bool) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "apply"}
	cmd.Flags().StringP("file", "f", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	cmd.Flags().Bool("rollback-on-failure", false, "")
	require.NoError(t, cmd.Flags().Set("file", file))
	if dryRun {
		require.NoError(t, cmd.Flags().Set("dry-run", "true"))
	}
	if yes {
		require.NoError(t, cmd.Flags().Set("yes", "true"))
	}
	if rollback {
		require.NoError(t, cmd.Flags().Set("rollback-on-failure", "true"))
	}
	return cmd
}

// writeApplyConfig writes a YAML config to a temp file and returns its path.
func writeApplyConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "apply.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

// portsByPrefix lists staging ports whose name begins with prefix.
func portsByPrefix(t *testing.T, prefix string) []*megaport.Port {
	t.Helper()
	client := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	all, err := client.PortService.ListPorts(ctx)
	require.NoError(t, err, "SDK ListPorts failed")
	var out []*megaport.Port
	for _, p := range all {
		if p != nil && strings.HasPrefix(p.Name, prefix) {
			out = append(out, p)
		}
	}
	return out
}

// portByExactName returns the port with the given exact name from the set
// matching prefix, or nil if not present.
func portByExactName(t *testing.T, prefix, name string) *megaport.Port {
	t.Helper()
	for _, p := range portsByPrefix(t, prefix) {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// portFromSDK reads a single port via GetPort (which populates AssociatedVXCs).
func portFromSDK(t *testing.T, uid string) *megaport.Port {
	t.Helper()
	client := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, err := client.PortService.GetPort(ctx, uid)
	require.NoErrorf(t, err, "SDK GetPort failed for %s", uid)
	require.NotNilf(t, p, "SDK GetPort returned nil for %s", uid)
	return p
}

// waitForDecommission polls status until it contains "DECOMMISSION" (covering
// DECOMMISSIONING and DECOMMISSIONED), the resource is gone, or the timeout
// elapses. Best-effort: logs and returns rather than failing, so a slow
// teardown does not mask the test result.
func waitForDecommission(t *testing.T, kind, uid string, status func(context.Context) (string, error)) {
	t.Helper()
	deadline := time.Now().Add(cleanupStatusTimeout)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s, err := status(ctx)
		cancel()
		if err != nil {
			t.Logf("cleanup: get %s %s after delete: %v (may already be gone)", kind, uid, err)
			return
		}
		if strings.Contains(s, "DECOMMISSION") {
			return
		}
		if time.Now().After(deadline) {
			t.Logf("cleanup: %s %s did not reach DECOMMISSION within %s (last status %q)", kind, uid, cleanupStatusTimeout, s)
			return
		}
		time.Sleep(cleanupStatusInterval)
	}
}

// registerSweepCleanup schedules a best-effort teardown of every port whose name
// begins with prefix and any VXC attached to those ports. VXCs are deleted
// first and polled to DECOMMISSIONING before ports are deleted, because staging
// rejects deleting a port with a live VXC attached. Runs even when the test
// fails, preventing orphaned billing resources.
func registerSweepCleanup(t *testing.T, prefix string) {
	t.Helper()
	t.Cleanup(func() {
		client := testutil.SharedIntegrationClient(t)

		listCtx, listCancel := context.WithTimeout(context.Background(), 60*time.Second)
		all, err := client.PortService.ListPorts(listCtx)
		listCancel()
		if err != nil {
			t.Logf("cleanup: ListPorts failed, resources may be orphaned: %v", err)
			return
		}
		var matched []*megaport.Port
		for _, p := range all {
			if p != nil && strings.HasPrefix(p.Name, prefix) {
				matched = append(matched, p)
			}
		}

		// 1. Delete attached VXCs first (dedup across both ports of a VXC).
		seenVXC := map[string]bool{}
		for _, p := range matched {
			getCtx, getCancel := context.WithTimeout(context.Background(), 30*time.Second)
			full, err := client.PortService.GetPort(getCtx, p.UID)
			getCancel()
			if err != nil || full == nil {
				t.Logf("cleanup: GetPort %s: %v (skipping VXC collection)", p.UID, err)
				continue
			}
			for _, vxc := range full.AssociatedVXCs {
				if vxc == nil || vxc.UID == "" || seenVXC[vxc.UID] {
					continue
				}
				seenVXC[vxc.UID] = true
				delCtx, delCancel := context.WithTimeout(context.Background(), 30*time.Second)
				err := client.VXCService.DeleteVXC(delCtx, vxc.UID, &megaport.DeleteVXCRequest{DeleteNow: true})
				delCancel()
				if err != nil {
					t.Logf("cleanup: delete VXC %s: %v", vxc.UID, err)
				}
			}
		}
		for uid := range seenVXC {
			waitForDecommission(t, "VXC", uid, func(ctx context.Context) (string, error) {
				vxc, err := client.VXCService.GetVXC(ctx, uid)
				if err != nil {
					return "", err
				}
				if vxc == nil {
					return "DECOMMISSIONED", nil
				}
				return vxc.ProvisioningStatus, nil
			})
		}

		// 2. Delete ports.
		for _, p := range matched {
			delCtx, delCancel := context.WithTimeout(context.Background(), 30*time.Second)
			_, err := client.PortService.DeletePort(delCtx, &megaport.DeletePortRequest{PortID: p.UID, DeleteNow: true})
			delCancel()
			if err != nil {
				t.Logf("cleanup: delete port %s: %v", p.UID, err)
			}
		}
	})
}

// twoPortVXCConfig returns a YAML apply config with two ports and one
// port-to-port VXC whose endpoints reference the ports via {{.port.<name>}}
// templates. Names are derived from prefix so a single sweep cleans them up.
func twoPortVXCConfig(prefix string) (yamlContent, portAName, portBName, vxcName string) {
	portAName = prefix + "-PortA"
	portBName = prefix + "-PortB"
	vxcName = prefix + "-VXC"
	yamlContent = fmt.Sprintf(`ports:
  - name: %s
    location_id: %d
    speed: 1000
    term: 1
    marketplace_visibility: false
  - name: %s
    location_id: %d
    speed: 1000
    term: 1
    marketplace_visibility: false
vxcs:
  - name: %s
    rate_limit: 100
    term: 1
    a_end:
      product_uid: "{{.port.%s}}"
    b_end:
      product_uid: "{{.port.%s}}"
`, portAName, integrationLocationID, portBName, integrationLocationID, vxcName, portAName, portBName)
	return yamlContent, portAName, portBName, vxcName
}

func TestIntegration_ApplyDryRun(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)

	prefix := fmt.Sprintf("CLI-Apply-Test-%s", generateUniqueID(t))
	registerSweepCleanup(t, prefix) // safety net; dry-run should create nothing

	yamlContent, _, _, _ := twoPortVXCConfig(prefix)
	cfgPath := writeApplyConfig(t, yamlContent)
	cmd := applyIntegrationCmd(t, cfgPath, true /*dryRun*/, true /*yes*/, false /*rollback*/)

	out := output.CaptureOutput(func() {
		require.NoError(t, ApplyConfig(cmd, nil, true, "table"), "dry-run apply should succeed")
	})

	// Validation results are reported; the VXC's template endpoints cannot be
	// validated server-side without provisioning, so it is reported as skipped.
	// Table rendering may wrap long cell values, so assert on substrings that
	// survive wrapping rather than the full status string.
	assert.Contains(t, out, "valid", "dry-run should report port validation results")
	assert.NotContains(t, out, "invalid", "no resource should fail validation in this config")
	assert.Contains(t, out, "skipped: requires", "templated VXC should be skipped in dry-run")

	// Crucially, nothing was provisioned.
	created := portsByPrefix(t, prefix)
	assert.Emptyf(t, created, "dry-run must not provision any ports; found %d", len(created))
}
