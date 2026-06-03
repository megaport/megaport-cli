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

// integrationLocationID is the staging data centre used for all VXC lifecycle
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
// read the UID back from this map.
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
// flags pre-set. Tests call ports.BuyPort(buildPortCmd(...), nil, true).
func buildPortCmd(name string, locationID int) *cobra.Command {
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
	_ = cmd.Flags().Set("name", name)
	_ = cmd.Flags().Set("term", "1")
	_ = cmd.Flags().Set("port-speed", "1000")
	_ = cmd.Flags().Set("location-id", fmt.Sprintf("%d", locationID))
	_ = cmd.Flags().Set("marketplace-visibility", "false")
	_ = cmd.Flags().Set("yes", "true")
	return cmd
}

// buildVXCCmd constructs a cobra.Command for BuyVXC with the given name,
// endpoint UIDs, and rate limit pre-set. Tests call BuyVXC(buildVXCCmd(...), nil, true).
func buildVXCCmd(name, aEndUID, bEndUID string, rateLimit int) *cobra.Command {
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
	_ = cmd.Flags().Set("name", name)
	_ = cmd.Flags().Set("a-end-uid", aEndUID)
	_ = cmd.Flags().Set("b-end-uid", bEndUID)
	_ = cmd.Flags().Set("rate-limit", fmt.Sprintf("%d", rateLimit))
	_ = cmd.Flags().Set("term", "1")
	_ = cmd.Flags().Set("yes", "true")
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
// then finds the created port's UID by listing all ports via the SDK and
// matching by name. Spinner output goes to real stdout — interleaved output
// from parallel tests is harmless.
func buyPortAndGetUID(t *testing.T, portName string) string {
	t.Helper()
	cmd := buildPortCmd(portName, integrationLocationID)
	require.NoErrorf(t, ports.BuyPort(cmd, nil, true), "BuyPort failed for %q", portName)

	sdkClient := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	allPorts, err := sdkClient.PortService.ListPorts(ctx)
	require.NoErrorf(t, err, "ListPorts failed after creating port %q", portName)
	for _, p := range allPorts {
		if p != nil && p.Name == portName {
			return p.UID
		}
	}
	t.Fatalf("port %q not found in port list after creation", portName)
	return ""
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
func runBuyVXC(t *testing.T, cmd *cobra.Command, vxcName string) string {
	t.Helper()
	require.NoErrorf(t, BuyVXC(cmd, nil, true), "BuyVXC failed for %q", vxcName)
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

func TestIntegration_VXCPortToPortLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	portAName := fmt.Sprintf("CLI-Test-VXC-PortA-%s", id)
	portBName := fmt.Sprintf("CLI-Test-VXC-PortB-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-VXC-%s", id)

	portAUID := buyPortAndGetUID(t, portAName)
	registerPortCleanup(t, portAUID)
	t.Logf("Created port A: %s", portAUID)

	portBUID := buyPortAndGetUID(t, portBName)
	registerPortCleanup(t, portBUID)
	t.Logf("Created port B: %s", portBUID)

	vxcCmd := buildVXCCmd(vxcName, portAUID, portBUID, 100)
	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	registerVXCCleanup(t, vxcUID)
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
}

func TestIntegration_VXCVLANModificationLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	portAName := fmt.Sprintf("CLI-Test-VLAN-PortA-%s", id)
	portBName := fmt.Sprintf("CLI-Test-VLAN-PortB-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-VLAN-%s", id)

	portAUID := buyPortAndGetUID(t, portAName)
	registerPortCleanup(t, portAUID)

	portBUID := buyPortAndGetUID(t, portBName)
	registerPortCleanup(t, portBUID)

	vxcCmd := buildVXCCmd(vxcName, portAUID, portBUID, 100)
	require.NoError(t, vxcCmd.Flags().Set("a-end-vlan", "100"))
	require.NoError(t, vxcCmd.Flags().Set("b-end-vlan", "200"))
	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	registerVXCCleanup(t, vxcUID)
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
	registerPortCleanup(t, portAUID)

	portBUID := buyPortAndGetUID(t, portBName)
	registerPortCleanup(t, portBUID)

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

	vxcCmd := &cobra.Command{Use: "buy"}
	vxcCmd.Flags().Bool("interactive", false, "")
	vxcCmd.Flags().Bool("no-wait", false, "")
	vxcCmd.Flags().Bool("yes", false, "")
	vxcCmd.Flags().String("json", "", "")
	vxcCmd.Flags().String("json-file", "", "")
	vxcCmd.Flags().String("name", "", "")
	vxcCmd.Flags().String("a-end-uid", "", "")
	vxcCmd.Flags().String("b-end-uid", "", "")
	vxcCmd.Flags().Int("rate-limit", 0, "")
	vxcCmd.Flags().Int("term", 0, "")
	vxcCmd.Flags().Int("a-end-vlan", 0, "")
	vxcCmd.Flags().Int("b-end-vlan", 0, "")
	vxcCmd.Flags().Int("a-end-inner-vlan", 0, "")
	vxcCmd.Flags().Int("b-end-inner-vlan", 0, "")
	vxcCmd.Flags().Int("a-end-vnic-index", 0, "")
	vxcCmd.Flags().Int("b-end-vnic-index", 0, "")
	vxcCmd.Flags().String("promo-code", "", "")
	vxcCmd.Flags().String("service-key", "", "")
	vxcCmd.Flags().String("cost-centre", "", "")
	vxcCmd.Flags().String("a-end-partner-config", "", "")
	vxcCmd.Flags().String("b-end-partner-config", "", "")
	require.NoError(t, vxcCmd.Flags().Set("json", string(buyJSON)))

	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	registerVXCCleanup(t, vxcUID)
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
