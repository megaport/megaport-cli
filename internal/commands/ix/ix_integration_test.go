//go:build integration

package ix

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateUniqueID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

// extractCreatedUID parses "<ResourceType> created <uid>" from buy command output.
func extractCreatedUID(captured, resourceType string) (string, bool) {
	prefix := resourceType + " created "
	for _, line := range strings.Split(captured, "\n") {
		if idx := strings.Index(line, prefix); idx >= 0 {
			uid := strings.TrimSpace(line[idx+len(prefix):])
			if uid != "" {
				return uid, true
			}
		}
	}
	return "", false
}

func integrationListIXCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("asn", 0, "")
	cmd.Flags().Int("vlan", 0, "")
	cmd.Flags().String("network-service-type", "", "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	return cmd
}

func integrationBuyIXCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().BoolP("interactive", "i", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("network-service-type", "", "")
	cmd.Flags().Int("asn", 0, "")
	cmd.Flags().String("mac-address", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().Int("vlan", 0, "")
	cmd.Flags().Bool("shutdown", false, "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationDeleteIXCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("later", false, "")
	return cmd
}

func integrationUpdateIXCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().BoolP("interactive", "i", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Int("vlan", 0, "")
	cmd.Flags().String("mac-address", "", "")
	cmd.Flags().Int("asn", 0, "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().Bool("public-graph", false, "")
	cmd.Flags().String("reverse-dns", "", "")
	cmd.Flags().String("a-end-product-uid", "", "")
	cmd.Flags().Bool("shutdown", false, "")
	return cmd
}

func integrationBuyPortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().BoolP("interactive", "i", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("port-speed", 0, "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationDeletePortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	return cmd
}

func integrationGetIXCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("export", false, "")
	return cmd
}

func integrationStatusIXCmd() *cobra.Command {
	return &cobra.Command{Use: "status"}
}

// TestIntegration_IXListAndGet is a fast read-only smoke test against staging.
func TestIntegration_IXListAndGet(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()
	t.Cleanup(func() { output.SetOutputFormat("table") })

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListIXs(integrationListIXCmd(), nil, true, "json")
	})
	require.NoError(t, listErr)

	var ixs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &ixs), "ListIXs returned invalid JSON")
	if len(ixs) == 0 {
		t.Skip("no IXs available on staging to test Get")
	}

	first := ixs[0]
	assert.Contains(t, first, "uid", "IX should have a uid field")
	assert.Contains(t, first, "name", "IX should have a name field")

	uid, ok := first["uid"].(string)
	require.True(t, ok && uid != "", "uid must be a non-empty string")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetIX(integrationGetIXCmd(), []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var gotList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &gotList), "GetIX JSON output must be valid")
	require.Len(t, gotList, 1)
	gotIX := gotList[0]
	assert.Equal(t, uid, gotIX["uid"])
	assert.Contains(t, gotIX, "name")
	assert.Contains(t, gotIX, "status")
}

// TestIntegration_IXLifecycle exercises the full create → get → status → update → delete path.
// It skips when no active IXs exist on staging (needed to discover a valid network-service-type).
// Expected runtime: up to ~25 minutes (two provisioning waits of up to 10 min each).
func TestIntegration_IXLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()
	// Ensure table format so PrintResourceCreated writes to stdout, not stderr.
	output.SetOutputFormat("table")
	t.Cleanup(func() { output.SetOutputFormat("table") })

	// Discover a valid IX network-service-type and location from existing active IXs.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	existingIXs, err := client.IXService.ListIXs(ctx, &megaport.ListIXsRequest{IncludeInactive: false})
	require.NoError(t, err, "failed to list IXs for type discovery")
	var firstIX *megaport.IX
	for _, ix := range existingIXs {
		if ix != nil {
			firstIX = ix
			break
		}
	}
	if firstIX == nil {
		t.Skip("no usable active IXs on staging — cannot determine a valid network-service-type for lifecycle test")
	}
	networkServiceType := firstIX.NetworkServiceType
	locationID := firstIX.LocationID

	// Create the port that will serve as the A-end of the IX.
	portName := fmt.Sprintf("CLI-Test-Port-IX-%s", generateUniqueID())
	pCmd := integrationBuyPortCmd()
	require.NoError(t, pCmd.Flags().Set("name", portName))
	require.NoError(t, pCmd.Flags().Set("term", "1"))
	require.NoError(t, pCmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, pCmd.Flags().Set("location-id", fmt.Sprintf("%d", locationID)))
	require.NoError(t, pCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, pCmd.Flags().Set("yes", "true"))

	var portErr error
	portOut := output.CaptureOutput(func() {
		portErr = ports.BuyPort(pCmd, nil, true)
	})
	require.NoError(t, portErr, "failed to create test port")

	portUID, ok := extractCreatedUID(portOut, "Port")
	require.True(t, ok, "could not extract port UID from output:\n%s", portOut)

	// Port cleanup runs last (registered first — t.Cleanup is LIFO).
	t.Cleanup(func() {
		output.SetOutputFormat("table")
		delPortCmd := integrationDeletePortCmd()
		_ = delPortCmd.Flags().Set("force", "true")
		output.CaptureOutput(func() {
			if err := ports.DeletePort(delPortCmd, []string{portUID}, true); err != nil {
				t.Errorf("cleanup: failed to delete test port %s: %v", portUID, err)
			}
		})
	})

	// Buy the IX.
	ixName := fmt.Sprintf("CLI-Test-IX-%s", generateUniqueID())
	buyCmd := integrationBuyIXCmd()
	require.NoError(t, buyCmd.Flags().Set("product-uid", portUID))
	require.NoError(t, buyCmd.Flags().Set("name", ixName))
	require.NoError(t, buyCmd.Flags().Set("network-service-type", networkServiceType))
	require.NoError(t, buyCmd.Flags().Set("asn", "65000"))
	require.NoError(t, buyCmd.Flags().Set("mac-address", "00:11:22:33:44:55"))
	require.NoError(t, buyCmd.Flags().Set("rate-limit", "1000"))
	require.NoError(t, buyCmd.Flags().Set("vlan", "100"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	var buyErr error
	buyOut := output.CaptureOutput(func() {
		buyErr = BuyIX(buyCmd, nil, true)
	})
	if buyErr != nil {
		errMsg := buyErr.Error()
		if strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "invalid") ||
			strings.Contains(errMsg, "not supported") {
			t.Skipf("IX type %q not available at location %d on staging: %v", networkServiceType, locationID, buyErr)
		}
		require.NoError(t, buyErr, "failed to create test IX")
	}

	ixUID, ok := extractCreatedUID(buyOut, "IX")
	require.True(t, ok, "could not extract IX UID from buy output:\n%s", buyOut)

	// IX cleanup runs first (registered second — LIFO).
	t.Cleanup(func() {
		output.SetOutputFormat("table")
		delCmd := integrationDeleteIXCmd()
		_ = delCmd.Flags().Set("force", "true")
		output.CaptureOutput(func() {
			if err := DeleteIX(delCmd, []string{ixUID}, true); err != nil {
				t.Errorf("cleanup: failed to delete test IX %s: %v", ixUID, err)
			}
		})
	})

	// Verify IX fields via GetIX.
	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetIX(integrationGetIXCmd(), []string{ixUID}, true, "json")
	})
	require.NoError(t, getErr)

	var gotList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &gotList))
	require.Len(t, gotList, 1)
	gotIX := gotList[0]
	assert.Equal(t, ixUID, gotIX["uid"])
	assert.Equal(t, ixName, gotIX["name"])
	assert.Contains(t, gotIX, "status")

	// Status as a sub-test to avoid an extra IX.
	t.Run("Status", func(t *testing.T) {
		var statusErr error
		statusOut := output.CaptureOutput(func() {
			statusErr = GetIXStatus(integrationStatusIXCmd(), []string{ixUID}, true, "json")
		})
		require.NoError(t, statusErr)

		var statusList []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(statusOut), &statusList))
		require.NotEmpty(t, statusList)
		assert.Equal(t, ixUID, statusList[0]["uid"])
		assert.Contains(t, statusList[0], "status")
	})

	// Update the IX name.
	updatedName := ixName + "-upd"
	updCmd := integrationUpdateIXCmd()
	require.NoError(t, updCmd.Flags().Set("name", updatedName))
	var updateErr error
	output.CaptureOutput(func() {
		updateErr = UpdateIX(updCmd, []string{ixUID}, true)
	})
	require.NoError(t, updateErr)

	// Verify the name was updated.
	var getErr2 error
	getOut2 := output.CaptureOutput(func() {
		getErr2 = GetIX(integrationGetIXCmd(), []string{ixUID}, true, "json")
	})
	require.NoError(t, getErr2)

	var gotList2 []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut2), &gotList2))
	require.Len(t, gotList2, 1)
	assert.Equal(t, updatedName, gotList2[0]["name"])
}
