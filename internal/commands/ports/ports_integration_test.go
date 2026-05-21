//go:build integration

package ports

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationLocationID is the staging data center used for port lifecycle
// tests. ID 67 is the canonical example location across the CLI's README,
// example flag strings, and the previous binary-invocation integration suite.
// It is stable on staging and supports the 1G/10G port speeds these tests
// exercise.
const integrationLocationID = 67

func generateUniqueID() string {
	buf := make([]byte, 4)
	if _, err := crypto_rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

// extractCreatedUID parses "<resource> created <uid>" from captured output and
// returns the UID. PrintResourceCreated writes "<resource> created <uid>" via
// PrintSuccess; with noColor=true the UID is unstyled, so the next whitespace
// character terminates it.
func extractCreatedUID(t *testing.T, captured, resourceLabel string) string {
	t.Helper()
	marker := resourceLabel + " created "
	idx := strings.Index(captured, marker)
	require.GreaterOrEqualf(t, idx, 0, "expected %q in output, got: %q", marker, captured)
	rest := captured[idx+len(marker):]
	end := strings.IndexAny(rest, " \n\r\t")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func newBuyPortCmd() *cobra.Command {
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
	return cmd
}

func newBuyLAGPortCmd() *cobra.Command {
	cmd := newBuyPortCmd()
	cmd.Flags().Int("lag-count", 0, "")
	return cmd
}

func newUpdatePortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Int("term", 0, "")
	return cmd
}

func newDeletePortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("now", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	return cmd
}

func newGetPortCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("watch", false, "")
	cmd.Flags().Bool("export", false, "")
	return cmd
}

// captureWithFormat sets the global output format and runs f under
// output.CaptureOutput. Both happen while CaptureOutput's stdoutMu is held, so
// parallel tests cannot race on the global format or on stdout swapping.
func captureWithFormat(format string, f func()) string {
	return output.CaptureOutput(func() {
		output.SetOutputFormat(format)
		f()
	})
}

// registerPortCleanup schedules a best-effort delete of the given port. The
// cleanup runs even when the test fails, ensuring no orphaned resources on
// staging. After deletion it asserts the port is reported as
// DECOMMISSIONING / DECOMMISSIONED.
//
// The cleanup re-installs the staging client into config.LoginFunc for the
// duration of its run. This is necessary because the outer
// `defer testutil.LoginWithClient(t, client)()` in the test body runs (and
// restores the original loginFunc) before any t.Cleanup callback, so cleanup
// would otherwise hit the real loginFunc — which may resolve to a different
// environment than staging.
func registerPortCleanup(t *testing.T, client *megaport.Client, uid string) {
	t.Helper()
	t.Cleanup(func() {
		restore := testutil.LoginWithClient(t, client)
		defer restore()

		delCmd := newDeletePortCmd()
		require.NoError(t, delCmd.Flags().Set("now", "true"))
		require.NoError(t, delCmd.Flags().Set("force", "true"))

		_ = captureWithFormat("table", func() {
			if err := DeletePort(delCmd, []string{uid}, true); err != nil {
				t.Errorf("cleanup: failed to delete port %s: %v", uid, err)
			}
		})

		getCmd := newGetPortCmd()
		var getErr error
		getOut := captureWithFormat("json", func() {
			getErr = GetPort(getCmd, []string{uid}, true, "json")
		})
		if getErr != nil {
			t.Logf("cleanup: GetPort after delete returned %v (port may already be gone)", getErr)
			return
		}
		var ports []map[string]interface{}
		if err := json.Unmarshal([]byte(getOut), &ports); err != nil || len(ports) == 0 {
			t.Logf("cleanup: could not parse GetPort JSON: %v, output: %s", err, getOut)
			return
		}
		status, _ := ports[0]["provisioning_status"].(string)
		if !strings.Contains(status, "DECOMMISSIONING") && !strings.Contains(status, "DECOMMISSIONED") {
			t.Errorf("expected port %s to be DECOMMISSIONING or DECOMMISSIONED, got %q", uid, status)
		}
	})
}

// portFromGet runs GetPort with JSON output and returns the first port record.
// Fails the test if the response is empty or malformed.
func portFromGet(t *testing.T, uid string) map[string]interface{} {
	t.Helper()
	cmd := newGetPortCmd()
	var err error
	captured := captureWithFormat("json", func() {
		err = GetPort(cmd, []string{uid}, true, "json")
	})
	require.NoError(t, err)
	var ports []map[string]interface{}
	require.NoErrorf(t, json.Unmarshal([]byte(captured), &ports), "GetPort output should be valid JSON: %s", captured)
	require.NotEmptyf(t, ports, "GetPort returned empty array: %s", captured)
	return ports[0]
}

func runBuyPort(t *testing.T, cmd *cobra.Command, resourceLabel string) string {
	t.Helper()
	var err error
	captured := captureWithFormat("table", func() {
		err = BuyPort(cmd, nil, true)
	})
	require.NoErrorf(t, err, "BuyPort failed: %s", captured)
	uid := extractCreatedUID(t, captured, resourceLabel)
	require.NotEmpty(t, uid, "extracted UID is empty")
	return uid
}

func runBuyLAGPort(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	var err error
	captured := captureWithFormat("table", func() {
		err = BuyLAGPort(cmd, nil, true)
	})
	require.NoErrorf(t, err, "BuyLAGPort failed: %s", captured)
	uid := extractCreatedUID(t, captured, "LAG Port")
	require.NotEmpty(t, uid, "extracted LAG UID is empty")
	return uid
}

func runUpdatePortName(t *testing.T, uid, newName string) {
	t.Helper()
	cmd := newUpdatePortCmd()
	require.NoError(t, cmd.Flags().Set("name", newName))

	var err error
	captured := captureWithFormat("table", func() {
		err = UpdatePort(cmd, []string{uid}, true)
	})
	require.NoErrorf(t, err, "UpdatePort failed: %s", captured)
}

func runUpdatePortJSON(t *testing.T, uid, jsonPayload string) {
	t.Helper()
	cmd := newUpdatePortCmd()
	require.NoError(t, cmd.Flags().Set("json", jsonPayload))

	var err error
	captured := captureWithFormat("table", func() {
		err = UpdatePort(cmd, []string{uid}, true)
	})
	require.NoErrorf(t, err, "UpdatePort failed: %s", captured)
}

func runUpdatePortJSONFile(t *testing.T, uid, jsonFile string) {
	t.Helper()
	cmd := newUpdatePortCmd()
	require.NoError(t, cmd.Flags().Set("json-file", jsonFile))

	var err error
	captured := captureWithFormat("table", func() {
		err = UpdatePort(cmd, []string{uid}, true)
	})
	require.NoErrorf(t, err, "UpdatePort failed: %s", captured)
}

func TestIntegration_PortLifecycle(t *testing.T) {
	t.Parallel()

	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	portName := fmt.Sprintf("CLI-Test-Port-%s", generateUniqueID())

	buyCmd := newBuyPortCmd()
	require.NoError(t, buyCmd.Flags().Set("name", portName))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, buyCmd.Flags().Set("location-id", fmt.Sprintf("%d", integrationLocationID)))
	require.NoError(t, buyCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	uid := runBuyPort(t, buyCmd, "Port")
	registerPortCleanup(t, client, uid)
	t.Logf("Created port with UID: %s", uid)

	port := portFromGet(t, uid)
	assert.Equal(t, uid, port["uid"])
	assert.Equal(t, portName, port["name"])
	assert.NotEmpty(t, port["provisioning_status"], "provisioning_status should be populated")

	newName := portName + "-Updated"
	runUpdatePortName(t, uid, newName)

	updated := portFromGet(t, uid)
	assert.Equal(t, newName, updated["name"])
}

func TestIntegration_LAGPortLifecycle(t *testing.T) {
	t.Parallel()

	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	portName := fmt.Sprintf("CLI-Test-LAG-%s", generateUniqueID())

	buyCmd := newBuyLAGPortCmd()
	require.NoError(t, buyCmd.Flags().Set("name", portName))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("port-speed", "10000"))
	require.NoError(t, buyCmd.Flags().Set("location-id", fmt.Sprintf("%d", integrationLocationID)))
	require.NoError(t, buyCmd.Flags().Set("lag-count", "1"))
	require.NoError(t, buyCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	uid := runBuyLAGPort(t, buyCmd)
	registerPortCleanup(t, client, uid)
	t.Logf("Created LAG port with UID: %s", uid)

	port := portFromGet(t, uid)
	assert.Equal(t, uid, port["uid"])
	assert.Equal(t, portName, port["name"])
	assert.NotEmpty(t, port["provisioning_status"])

	newName := portName + "-Updated"
	runUpdatePortName(t, uid, newName)

	updated := portFromGet(t, uid)
	assert.Equal(t, newName, updated["name"])
}

func TestIntegration_PortJSONInputLifecycle(t *testing.T) {
	t.Parallel()

	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	portName := fmt.Sprintf("CLI-Test-Port-JSON-%s", generateUniqueID())

	buyPayload := map[string]interface{}{
		"name":                  portName,
		"term":                  1,
		"portSpeed":             1000,
		"locationId":            integrationLocationID,
		"marketPlaceVisibility": false,
	}
	buyJSON, err := json.Marshal(buyPayload)
	require.NoError(t, err)

	buyCmd := newBuyPortCmd()
	require.NoError(t, buyCmd.Flags().Set("json", string(buyJSON)))

	uid := runBuyPort(t, buyCmd, "Port")
	registerPortCleanup(t, client, uid)
	t.Logf("Created port (JSON input) with UID: %s", uid)

	port := portFromGet(t, uid)
	assert.Equal(t, uid, port["uid"])
	assert.Equal(t, portName, port["name"])
	assert.NotEmpty(t, port["provisioning_status"])

	newName := portName + "-Updated-JSON"
	updatePayload, err := json.Marshal(map[string]string{"name": newName})
	require.NoError(t, err)

	runUpdatePortJSON(t, uid, string(updatePayload))

	updated := portFromGet(t, uid)
	assert.Equal(t, newName, updated["name"])
}

func TestIntegration_PortJSONFileLifecycle(t *testing.T) {
	t.Parallel()

	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	portName := fmt.Sprintf("CLI-Test-Port-JSONFile-%s", generateUniqueID())

	buyPayload := map[string]interface{}{
		"name":                  portName,
		"term":                  1,
		"portSpeed":             1000,
		"locationId":            integrationLocationID,
		"marketPlaceVisibility": false,
	}
	buyJSON, err := json.MarshalIndent(buyPayload, "", "  ")
	require.NoError(t, err)

	buyFile := filepath.Join(t.TempDir(), "port-buy.json")
	require.NoError(t, os.WriteFile(buyFile, buyJSON, 0o600))

	buyCmd := newBuyPortCmd()
	require.NoError(t, buyCmd.Flags().Set("json-file", buyFile))

	uid := runBuyPort(t, buyCmd, "Port")
	registerPortCleanup(t, client, uid)
	t.Logf("Created port (JSON file) with UID: %s", uid)

	port := portFromGet(t, uid)
	assert.Equal(t, uid, port["uid"])
	assert.Equal(t, portName, port["name"])
	assert.NotEmpty(t, port["provisioning_status"])

	newName := portName + "-Updated-TempJSON"
	updatePayload, err := json.MarshalIndent(map[string]interface{}{
		"name":                  newName,
		"marketPlaceVisibility": true,
	}, "", "  ")
	require.NoError(t, err)

	updateFile := filepath.Join(t.TempDir(), "port-update.json")
	require.NoError(t, os.WriteFile(updateFile, updatePayload, 0o600))

	runUpdatePortJSONFile(t, uid, updateFile)

	updated := portFromGet(t, uid)
	assert.Equal(t, newName, updated["name"])
}
