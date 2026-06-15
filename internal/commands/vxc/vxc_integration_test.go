//go:build integration

package vxc

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationLocationID is the staging data center used for all VXC lifecycle
// tests. Location 67 matches the canonical example used across the CLI test
// suite and supports the 1G port speed these tests exercise.
const integrationLocationID = 67

const (
	cleanupStatusTimeout  = 60 * time.Second
	cleanupStatusInterval = 2 * time.Second
)

// integrationVXCBuyResponses lets parallel tests retrieve the BuyVXCResponse
// for their VXC without scraping stdout. The init() hook below wraps
// buyVXCFunc so the SDK response is stored under the request VXCName; tests
// read the UID back from this map. key: req.VXCName (string), value: *megaport.BuyVXCResponse
var integrationVXCBuyResponses sync.Map

func init() {
	base := buyVXCFunc
	buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
		resp, err := base(ctx, client, req)
		if err == nil && resp != nil && req != nil && req.VXCName != "" {
			integrationVXCBuyResponses.Store(req.VXCName, resp)
		}
		return resp, err
	}
}

func generateUniqueID(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 4)
	_, err := crypto_rand.Read(buf)
	require.NoError(t, err, "failed to read crypto/rand entropy")
	return hex.EncodeToString(buf)
}

// buildPortCmd constructs a cobra.Command for ports.BuyPort with all required
// flags pre-set. Tests call ports.BuyPort(buildPortCmd(t, ...), nil, true).
func buildPortCmd(t *testing.T, name string, locationID int) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("port-speed", 0, "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("resource-tags", "", "")
	cmd.Flags().Bool("cost-confirm", true, "")
	require.NoError(t, cmd.Flags().Set("name", name))
	require.NoError(t, cmd.Flags().Set("term", "1"))
	require.NoError(t, cmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, cmd.Flags().Set("location-id", fmt.Sprintf("%d", locationID)))
	require.NoError(t, cmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	return cmd
}

// buildVXCCmd constructs a cobra.Command for BuyVXC with the given name,
// endpoint UIDs, and rate limit pre-set. Tests call BuyVXC(buildVXCCmd(t, ...), nil, true).
func buildVXCCmd(t *testing.T, name, aEndUID, bEndUID string, rateLimit int) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("a-end-uid", "", "")
	cmd.Flags().String("b-end-uid", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("a-end-vlan", 0, "")
	cmd.Flags().Int("b-end-vlan", 0, "")
	cmd.Flags().Int("a-end-inner-vlan", 0, "")
	cmd.Flags().Int("b-end-inner-vlan", 0, "")
	cmd.Flags().Int("a-end-vnic-index", 0, "")
	cmd.Flags().Int("b-end-vnic-index", 0, "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("service-key", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("a-end-partner-config", "", "")
	cmd.Flags().String("b-end-partner-config", "", "")
	require.NoError(t, cmd.Flags().Set("name", name))
	require.NoError(t, cmd.Flags().Set("a-end-uid", aEndUID))
	require.NoError(t, cmd.Flags().Set("b-end-uid", bEndUID))
	require.NoError(t, cmd.Flags().Set("rate-limit", fmt.Sprintf("%d", rateLimit)))
	require.NoError(t, cmd.Flags().Set("term", "1"))
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	return cmd
}

func newDeletePortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("now", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	return cmd
}

func newDeleteVXCCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("later", false, "")
	return cmd
}

func newUpdateVXCCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().Int("a-end-vlan", 0, "")
	cmd.Flags().Int("b-end-vlan", 0, "")
	cmd.Flags().String("a-end-location", "", "")
	cmd.Flags().String("b-end-location", "", "")
	cmd.Flags().Bool("locked", false, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Bool("shutdown", false, "")
	cmd.Flags().Int("a-end-inner-vlan", 0, "")
	cmd.Flags().Int("b-end-inner-vlan", 0, "")
	cmd.Flags().String("a-end-uid", "", "")
	cmd.Flags().String("b-end-uid", "", "")
	cmd.Flags().String("a-end-partner-config", "", "")
	cmd.Flags().String("b-end-partner-config", "", "")
	cmd.Flags().Bool("is-approved", false, "")
	cmd.Flags().Int("a-vnic-index", -1, "")
	cmd.Flags().Int("b-vnic-index", -1, "")
	return cmd
}

// buyPortAndGetUID calls ports.BuyPort (which blocks until the port is LIVE)
// then recovers the created port's UID from the BuyPortResponse captured by
// the ports package's integration buy hook. Cleanup is registered immediately
// after the buy succeeds and before the UID is asserted, so a billable port is
// never leaked if UID recovery fails.
func buyPortAndGetUID(t *testing.T, portName string) string {
	t.Helper()
	cmd := buildPortCmd(t, portName, integrationLocationID)
	require.NoErrorf(t, ports.BuyPort(cmd, nil, true), "BuyPort failed for %q", portName)

	uid, ok := ports.IntegrationBuyPortUID(portName)
	require.Truef(t, ok, "no port buy response captured for %q", portName)
	registerPortCleanup(t, uid)
	return uid
}

// vxcBuyResponseUID returns the TechnicalServiceUID captured by the init() hook
// for the given VXC name without failing the test. ok is false if no response
// was recorded or it carried no UID. Used to register cleanup before the
// fail-able UID assertions in uidFromVXCBuyResponse.
func vxcBuyResponseUID(vxcName string) (uid string, ok bool) {
	v, loaded := integrationVXCBuyResponses.Load(vxcName)
	if !loaded {
		return "", false
	}
	resp, isResp := v.(*megaport.BuyVXCResponse)
	if !isResp || resp.TechnicalServiceUID == "" {
		return "", false
	}
	return resp.TechnicalServiceUID, true
}

// uidFromVXCBuyResponse returns the TechnicalServiceUID captured by the
// init() hook for the given VXC name.
func uidFromVXCBuyResponse(t *testing.T, vxcName string) string {
	t.Helper()
	v, ok := integrationVXCBuyResponses.Load(vxcName)
	require.Truef(t, ok, "no VXC buy response recorded for %q", vxcName)
	resp, ok := v.(*megaport.BuyVXCResponse)
	require.Truef(t, ok, "VXC buy response for %q has unexpected type %T", vxcName, v)
	require.NotEmptyf(t, resp.TechnicalServiceUID, "no UID in VXC buy response for %q", vxcName)
	return resp.TechnicalServiceUID
}

// runBuyVXC calls BuyVXC and returns the UID from the hook-captured response.
// Cleanup is registered immediately after the buy succeeds and before the UID
// is asserted, so a billable VXC is never leaked if UID extraction fails.
func runBuyVXC(t *testing.T, cmd *cobra.Command, vxcName string) string {
	t.Helper()
	require.NoErrorf(t, BuyVXC(cmd, nil, true), "BuyVXC failed for %q", vxcName)
	if uid, ok := vxcBuyResponseUID(vxcName); ok {
		registerVXCCleanup(t, uid)
	}
	return uidFromVXCBuyResponse(t, vxcName)
}

// registerPortCleanup schedules a best-effort delete of the port. Register
// before registering VXC cleanup so port deletions run after VXC deletion
// (t.Cleanup is LIFO).
func registerPortCleanup(t *testing.T, uid string) {
	t.Helper()
	t.Cleanup(func() {
		delCmd := newDeletePortCmd()
		require.NoError(t, delCmd.Flags().Set("now", "true"))
		require.NoError(t, delCmd.Flags().Set("force", "true"))
		if err := ports.DeletePort(delCmd, []string{uid}, true); err != nil {
			t.Errorf("cleanup: failed to delete port %s: %v", uid, err)
			return
		}
	})
}

// registerVXCCleanup schedules VXC deletion and polls for DECOMMISSIONING so
// port cleanups (registered before this) can safely proceed.
func registerVXCCleanup(t *testing.T, uid string) {
	t.Helper()
	t.Cleanup(func() {
		delCmd := newDeleteVXCCmd()
		require.NoError(t, delCmd.Flags().Set("force", "true"))
		if err := DeleteVXC(delCmd, []string{uid}, true); err != nil {
			t.Errorf("cleanup: failed to delete VXC %s: %v", uid, err)
			return
		}

		sdkClient := testutil.SharedIntegrationClient(t)
		deadline := time.Now().Add(cleanupStatusTimeout)
		var lastStatus string
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			vxc, err := sdkClient.VXCService.GetVXC(ctx, uid)
			cancel()
			if err != nil {
				t.Logf("cleanup: GetVXC after delete returned %v (may already be gone)", err)
				return
			}
			if vxc == nil {
				return
			}
			lastStatus = vxc.ProvisioningStatus
			if strings.Contains(lastStatus, "DECOMMISSIONING") || strings.Contains(lastStatus, "DECOMMISSIONED") {
				return
			}
			if time.Now().After(deadline) {
				t.Errorf("VXC %s did not reach DECOMMISSIONING within %s, last status %q", uid, cleanupStatusTimeout, lastStatus)
				return
			}
			time.Sleep(cleanupStatusInterval)
		}
	})
}

// vxcFromSDK reads VXC state via the shared SDK client. Used for state
// assertions in parallel tests instead of CaptureOutput on GetVXC.
func vxcFromSDK(t *testing.T, uid string) *megaport.VXC {
	t.Helper()
	sdkClient := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	vxc, err := sdkClient.VXCService.GetVXC(ctx, uid)
	require.NoErrorf(t, err, "SDK GetVXC failed for %s", uid)
	require.NotNilf(t, vxc, "SDK GetVXC returned nil for %s", uid)
	return vxc
}

func newUpdateVXCTagsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update-tags"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

// vxcTagsFromSDK reads a VXC's resource tags via the shared SDK client (the same
// call the list-tags command makes underneath). Used instead of capturing the
// command's stdout because these tests run under t.Parallel().
func vxcTagsFromSDK(t *testing.T, uid string) map[string]string {
	t.Helper()
	sdkClient := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tags, err := sdkClient.VXCService.ListVXCResourceTags(ctx, uid)
	require.NoErrorf(t, err, "SDK ListVXCResourceTags failed for %s", uid)
	return tags
}

// runVXCTagRoundTrip exercises update-tags then list-tags on an existing
// lifecycle VXC: it sets two tags via the update-tags command, reads them back
// through the SDK, then clears them and confirms they are gone. It rides on the
// VXC the caller already provisioned, so it needs no separate cleanup.
func runVXCTagRoundTrip(t *testing.T, uid string) {
	t.Helper()
	want := map[string]string{"env": "cli-integration", "owner": "esd-1392"}
	tagJSON, err := json.Marshal(want)
	require.NoError(t, err)

	setCmd := newUpdateVXCTagsCmd()
	require.NoError(t, setCmd.Flags().Set("json", string(tagJSON)))
	require.NoError(t, setCmd.Flags().Set("force", "true"))
	require.NoErrorf(t, UpdateVXCResourceTags(setCmd, []string{uid}, true), "UpdateVXCResourceTags failed for %s", uid)

	// Assert our tags round-tripped without requiring the map to contain only
	// them, so an API-injected tag can't make this flaky.
	got := vxcTagsFromSDK(t, uid)
	for k, v := range want {
		assert.Equalf(t, v, got[k], "tag %q should round-trip", k)
	}

	clearCmd := newUpdateVXCTagsCmd()
	require.NoError(t, clearCmd.Flags().Set("json", "{}"))
	require.NoError(t, clearCmd.Flags().Set("force", "true"))
	require.NoErrorf(t, UpdateVXCResourceTags(clearCmd, []string{uid}, true), "clearing VXC tags failed for %s", uid)

	cleared := vxcTagsFromSDK(t, uid)
	for k := range want {
		assert.NotContainsf(t, cleared, k, "tag %q should be cleared", k)
	}
}

func TestIntegration_VXCPortToPortLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	portAName := fmt.Sprintf("CLI-Test-VXC-PortA-%s", id)
	portBName := fmt.Sprintf("CLI-Test-VXC-PortB-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-VXC-%s", id)

	portAUID := buyPortAndGetUID(t, portAName)
	t.Logf("Created port A: %s", portAUID)

	portBUID := buyPortAndGetUID(t, portBName)
	t.Logf("Created port B: %s", portBUID)

	vxcCmd := buildVXCCmd(t, vxcName, portAUID, portBUID, 100)
	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	t.Logf("Created VXC: %s", vxcUID)

	vxc := vxcFromSDK(t, vxcUID)
	assert.Equal(t, vxcUID, vxc.UID)
	assert.Equal(t, vxcName, vxc.Name)
	assert.Equal(t, 100, vxc.RateLimit)
	assert.Equal(t, portAUID, vxc.AEndConfiguration.UID)
	assert.Equal(t, portBUID, vxc.BEndConfiguration.UID)

	newName := vxcName + "-Updated"
	updateCmd := newUpdateVXCCmd()
	require.NoError(t, updateCmd.Flags().Set("name", newName))
	require.NoErrorf(t, UpdateVXC(updateCmd, []string{vxcUID}, true), "UpdateVXC failed")

	updated := vxcFromSDK(t, vxcUID)
	assert.Equal(t, newName, updated.Name)

	runVXCTagRoundTrip(t, vxcUID)
}

func TestIntegration_VXCVLANModificationLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	portAName := fmt.Sprintf("CLI-Test-VLAN-PortA-%s", id)
	portBName := fmt.Sprintf("CLI-Test-VLAN-PortB-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-VLAN-%s", id)

	portAUID := buyPortAndGetUID(t, portAName)

	portBUID := buyPortAndGetUID(t, portBName)

	vxcCmd := buildVXCCmd(t, vxcName, portAUID, portBUID, 100)
	require.NoError(t, vxcCmd.Flags().Set("a-end-vlan", "100"))
	require.NoError(t, vxcCmd.Flags().Set("b-end-vlan", "200"))
	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	t.Logf("Created VXC: %s", vxcUID)

	vxc := vxcFromSDK(t, vxcUID)
	assert.Equal(t, 100, vxc.AEndConfiguration.VLAN)
	assert.Equal(t, 200, vxc.BEndConfiguration.VLAN)

	updateCmd := newUpdateVXCCmd()
	require.NoError(t, updateCmd.Flags().Set("a-end-vlan", "101"))
	require.NoError(t, updateCmd.Flags().Set("b-end-vlan", "201"))
	require.NoErrorf(t, UpdateVXC(updateCmd, []string{vxcUID}, true), "UpdateVXC VLAN update failed")

	updated := vxcFromSDK(t, vxcUID)
	assert.Equal(t, 101, updated.AEndConfiguration.VLAN)
	assert.Equal(t, 201, updated.BEndConfiguration.VLAN)
}

func TestIntegration_VXCJSONInputLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	portAName := fmt.Sprintf("CLI-Test-VXCJ-PortA-%s", id)
	portBName := fmt.Sprintf("CLI-Test-VXCJ-PortB-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-VXCJ-%s", id)

	portAUID := buyPortAndGetUID(t, portAName)

	portBUID := buyPortAndGetUID(t, portBName)

	buyPayload := map[string]any{
		"portUid":   portAUID,
		"vxcName":   vxcName,
		"rateLimit": 100,
		"term":      1,
		"bEndConfiguration": map[string]any{
			"productUID": portBUID,
		},
	}
	buyJSON, err := json.Marshal(buyPayload)
	require.NoError(t, err)

	vxcCmd := buildVXCCmd(t, "", "", "", 0)
	require.NoError(t, vxcCmd.Flags().Set("json", string(buyJSON)))

	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	t.Logf("Created VXC (JSON input): %s", vxcUID)

	vxc := vxcFromSDK(t, vxcUID)
	assert.Equal(t, vxcUID, vxc.UID)
	assert.Equal(t, vxcName, vxc.Name)
	assert.Equal(t, 100, vxc.RateLimit)

	newName := vxcName + "-Updated"
	updateCmd := newUpdateVXCCmd()
	require.NoError(t, updateCmd.Flags().Set("name", newName))
	require.NoErrorf(t, UpdateVXC(updateCmd, []string{vxcUID}, true), "UpdateVXC failed")

	updated := vxcFromSDK(t, vxcUID)
	assert.Equal(t, newName, updated.Name)
}
