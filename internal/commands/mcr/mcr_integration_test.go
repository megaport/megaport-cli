//go:build integration

package mcr

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stagingMCRLocationID is a known MCR-capable staging location. Discovering one
// via ListLocations would be overengineering for this lifecycle test.
const stagingMCRLocationID = 67

func generateUniqueID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	return hex.EncodeToString(buf)
}

func integrationMCRBuyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("port-speed", 0, "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Int("mcr-asn", 0, "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationMCRUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationMCRDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	return cmd
}

func integrationMCRPrefixFilterCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "prefix-filter-list"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("address-family", "", "")
	cmd.Flags().String("entries", "", "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

// parseCreatedUID pulls the resource UID out of a "<resource> created <uid>"
// success message.
func parseCreatedUID(out, resource string) string {
	marker := resource + " created "
	for _, line := range strings.Split(out, "\n") {
		if i := strings.Index(line, marker); i >= 0 {
			return strings.TrimSpace(line[i+len(marker):])
		}
	}
	return ""
}

// parsePrefixFilterListID pulls the numeric ID out of the
// "Prefix filter list created successfully - ID: <n>" success message.
func parsePrefixFilterListID(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "created successfully") {
			continue
		}
		if i := strings.Index(line, "ID:"); i >= 0 {
			return strings.TrimSpace(line[i+len("ID:"):])
		}
	}
	return ""
}

// getMCRJSON retrieves an MCR as JSON and returns the decoded object.
func getMCRJSON(t *testing.T, uid string) map[string]interface{} {
	t.Helper()
	var err error
	out := output.CaptureOutput(func() {
		err = GetMCR(&cobra.Command{Use: "get"}, []string{uid}, true, "json")
	})
	require.NoError(t, err, "get MCR output: %s", out)

	var mcrs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &mcrs), "MCR output should be valid JSON: %s", out)
	require.Len(t, mcrs, 1, "expected exactly one MCR")
	return mcrs[0]
}

func TestIntegration_MCRLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	name := fmt.Sprintf("CLI-Test-MCR-%s", generateUniqueID())

	// Buy a new MCR using flags. BuyMCR waits for provisioning (no --no-wait),
	// so the MCR is ready for prefix-filter-list operations once it returns.
	buyCmd := integrationMCRBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(stagingMCRLocationID)))
	require.NoError(t, buyCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	output.SetOutputFormat("table") // route the success message to stdout
	var buyErr error
	buyOut := output.CaptureOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut, "MCR")
	require.NotEmpty(t, mcrUID, "could not parse MCR UID from: %s", buyOut)

	t.Cleanup(func() {
		delCmd := integrationMCRDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		out := output.CaptureOutput(func() { _ = DeleteMCR(delCmd, []string{mcrUID}, true) })
		t.Logf("cleanup: delete MCR %s: %s", mcrUID, out)
	})

	// Get and verify the core fields are present.
	mcr := getMCRJSON(t, mcrUID)
	assert.Equal(t, mcrUID, mcr["uid"])
	assert.Equal(t, name, mcr["name"])
	assert.Contains(t, mcr, "provisioning_status")
	assert.NotEmpty(t, mcr["provisioning_status"])

	// Update the name and verify it took effect.
	newName := name + "-updated"
	updCmd := integrationMCRUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := output.CaptureOutput(func() { updErr = UpdateMCR(updCmd, []string{mcrUID}, true) })
	require.NoError(t, updErr, "update MCR output: %s", updOut)

	assert.Equal(t, newName, getMCRJSON(t, mcrUID)["name"])

	// Prefix filter list lifecycle (create -> get -> update -> delete).
	createPFLJSON := `{
		"description": "Test Prefix Filter List",
		"addressFamily": "IPv4",
		"entries": [
			{"action": "permit", "prefix": "10.0.1.0/24", "ge": 25, "le": 32},
			{"action": "deny", "prefix": "10.0.2.0/24", "ge": 0, "le": 25}
		]
	}`

	createCmd := integrationMCRPrefixFilterCmd()
	require.NoError(t, createCmd.Flags().Set("json", createPFLJSON))
	output.SetOutputFormat("table") // route the success message to stdout
	var createErr error
	createOut := output.CaptureOutput(func() {
		createErr = CreateMCRPrefixFilterList(createCmd, []string{mcrUID}, true)
	})
	require.NoError(t, createErr, "create prefix filter list output: %s", createOut)

	pflID := parsePrefixFilterListID(createOut)
	require.NotEmpty(t, pflID, "could not parse prefix filter list ID from: %s", createOut)

	// Best-effort cleanup. Registered after the MCR cleanup so it runs first
	// (cleanups run LIFO) — the list must be gone before the MCR is deleted.
	t.Cleanup(func() {
		out := output.CaptureOutput(func() {
			_ = DeleteMCRPrefixFilterList(&cobra.Command{Use: "delete"}, []string{mcrUID, pflID}, true)
		})
		t.Logf("cleanup: delete prefix filter list %s: %s", pflID, out)
	})

	// Get the prefix filter list and assert it exists with the expected content.
	var getPFLErr error
	getPFLOut := output.CaptureOutput(func() {
		getPFLErr = GetMCRPrefixFilterList(&cobra.Command{Use: "get"}, []string{mcrUID, pflID}, true, "json")
	})
	require.NoError(t, getPFLErr, "get prefix filter list output: %s", getPFLOut)
	assert.Contains(t, getPFLOut, "Test Prefix Filter List")
	assert.Contains(t, getPFLOut, "10.0.1.0/24")
	assert.Contains(t, getPFLOut, "10.0.2.0/24")

	// Update the prefix filter list, adding a new entry.
	updatePFLJSON := `{
		"description": "Test Prefix Filter List Updated",
		"addressFamily": "IPv4",
		"entries": [
			{"action": "permit", "prefix": "10.0.1.0/24", "ge": 25, "le": 32},
			{"action": "deny", "prefix": "10.0.2.0/24", "ge": 0, "le": 25},
			{"action": "permit", "prefix": "192.168.0.0/16", "ge": 24, "le": 32}
		]
	}`
	updatePFLCmd := integrationMCRPrefixFilterCmd()
	require.NoError(t, updatePFLCmd.Flags().Set("json", updatePFLJSON))
	var updatePFLErr error
	updatePFLOut := output.CaptureOutput(func() {
		updatePFLErr = UpdateMCRPrefixFilterList(updatePFLCmd, []string{mcrUID, pflID}, true)
	})
	require.NoError(t, updatePFLErr, "update prefix filter list output: %s", updatePFLOut)

	// Verify the update landed.
	var verifyPFLErr error
	verifyPFLOut := output.CaptureOutput(func() {
		verifyPFLErr = GetMCRPrefixFilterList(&cobra.Command{Use: "get"}, []string{mcrUID, pflID}, true, "json")
	})
	require.NoError(t, verifyPFLErr, "verify prefix filter list output: %s", verifyPFLOut)
	assert.Contains(t, verifyPFLOut, "Test Prefix Filter List Updated")
	assert.Contains(t, verifyPFLOut, "192.168.0.0/16")

	// Delete the prefix filter list and assert it is gone.
	output.SetOutputFormat("table")
	var deletePFLErr error
	deletePFLOut := output.CaptureOutput(func() {
		deletePFLErr = DeleteMCRPrefixFilterList(&cobra.Command{Use: "delete"}, []string{mcrUID, pflID}, true)
	})
	require.NoError(t, deletePFLErr, "delete prefix filter list output: %s", deletePFLOut)

	goneErr := GetMCRPrefixFilterList(&cobra.Command{Use: "get"}, []string{mcrUID, pflID}, true, "json")
	assert.Error(t, goneErr, "prefix filter list should no longer exist")
}

func TestIntegration_MCRJSONInputLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	name := fmt.Sprintf("CLI-JSON-MCR-%s", generateUniqueID())

	buyJSON := fmt.Sprintf(`{
		"name": "%s",
		"term": 1,
		"portSpeed": 1000,
		"locationId": %d,
		"marketplaceVisibility": false
	}`, name, stagingMCRLocationID)

	buyCmd := integrationMCRBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("json", buyJSON))

	output.SetOutputFormat("table")
	var buyErr error
	buyOut := output.CaptureOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR (JSON) output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut, "MCR")
	require.NotEmpty(t, mcrUID, "could not parse MCR UID from: %s", buyOut)

	t.Cleanup(func() {
		delCmd := integrationMCRDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		out := output.CaptureOutput(func() { _ = DeleteMCR(delCmd, []string{mcrUID}, true) })
		t.Logf("cleanup: delete MCR %s: %s", mcrUID, out)
	})

	mcr := getMCRJSON(t, mcrUID)
	assert.Equal(t, mcrUID, mcr["uid"])
	assert.Equal(t, name, mcr["name"])

	newName := name + "-updated"
	updCmd := integrationMCRUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := output.CaptureOutput(func() { updErr = UpdateMCR(updCmd, []string{mcrUID}, true) })
	require.NoError(t, updErr, "update MCR output: %s", updOut)

	assert.Equal(t, newName, getMCRJSON(t, mcrUID)["name"])
}
