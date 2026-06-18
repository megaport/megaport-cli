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

// captureTableOutput captures stdout with the output format forced to "table",
// since JSON mode routes success/info messages to stderr where CaptureOutput
// can't see them.
func captureTableOutput(f func()) string {
	output.SetOutputFormat("table")
	return output.CaptureOutput(f)
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

func integrationMCRTagCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update-tags"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

// tagsFromListJSON parses the JSON output of list-tags (an array of {key,value}
// objects) into a map, so assertions can check key->value pairs instead of
// substring-matching the rendered blob.
func tagsFromListJSON(t *testing.T, out string) map[string]string {
	t.Helper()
	var tags []output.ResourceTag
	require.NoErrorf(t, json.Unmarshal([]byte(out), &tags), "parse list-tags JSON: %s", out)
	m := make(map[string]string, len(tags))
	for _, tag := range tags {
		m[tag.Key] = tag.Value
	}
	return m
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
	testutil.RequireStagingForProvisioning(t)
	client := testutil.SetupIntegrationClient(t)
	// Restore login via t.Cleanup, not defer: defers run before t.Cleanup, so a
	// deferred restore would swap back the default login (wrong environment)
	// before the resource-deletion cleanups below get to run.
	t.Cleanup(testutil.LoginWithClient(t, client))

	// Action functions mutate the process-wide output format; restore it so
	// test order can't leak state between tests in this package.
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

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

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut, "MCR")
	// Register cleanup before asserting on mcrUID, so any created MCR is
	// deleted even if the UID parse fails.
	t.Cleanup(func() {
		if mcrUID == "" {
			return
		}
		delCmd := integrationMCRDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMCR(delCmd, []string{mcrUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MCR %s failed: %v; output: %s", mcrUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MCR %s: %s", mcrUID, out)
	})
	require.NotEmpty(t, mcrUID, "could not parse MCR UID from: %s", buyOut)

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
	updOut := captureTableOutput(func() { updErr = UpdateMCR(updCmd, []string{mcrUID}, true) })
	require.NoError(t, updErr, "update MCR output: %s", updOut)

	assert.Equal(t, newName, getMCRJSON(t, mcrUID)["name"])

	// Resource tag round-trip (ESD-1392): set tags via update-tags, read them
	// back via list-tags, then clear them. Rides on the lifecycle MCR, so no
	// extra cleanup is needed.
	want := map[string]string{"env": "cli-integration", "owner": "esd-1392"}
	setTagsJSON, err := json.Marshal(want)
	require.NoError(t, err)
	setTagsCmd := integrationMCRTagCmd()
	require.NoError(t, setTagsCmd.Flags().Set("json", string(setTagsJSON)))
	require.NoError(t, setTagsCmd.Flags().Set("force", "true"))
	var setTagsErr error
	setTagsOut := captureTableOutput(func() { setTagsErr = UpdateMCRResourceTags(setTagsCmd, []string{mcrUID}, true) })
	require.NoError(t, setTagsErr, "update MCR tags output: %s", setTagsOut)

	var listTagsErr error
	listTagsOut := output.CaptureOutput(func() {
		listTagsErr = ListMCRResourceTags(&cobra.Command{Use: "list-tags"}, []string{mcrUID}, true, "json")
	})
	require.NoError(t, listTagsErr, "list MCR tags output: %s", listTagsOut)
	// Assert our tags round-tripped without requiring the map to contain only
	// them, so an API-injected tag can't make this flaky.
	got := tagsFromListJSON(t, listTagsOut)
	for k, v := range want {
		assert.Equalf(t, v, got[k], "tag %q should round-trip", k)
	}

	// Clear the tags so the MCR is left clean for the steps that follow.
	clearTagsCmd := integrationMCRTagCmd()
	require.NoError(t, clearTagsCmd.Flags().Set("json", "{}"))
	require.NoError(t, clearTagsCmd.Flags().Set("force", "true"))
	var clearTagsErr error
	clearTagsOut := captureTableOutput(func() { clearTagsErr = UpdateMCRResourceTags(clearTagsCmd, []string{mcrUID}, true) })
	require.NoError(t, clearTagsErr, "clear MCR tags output: %s", clearTagsOut)

	var verifyTagsErr error
	verifyTagsOut := output.CaptureOutput(func() {
		verifyTagsErr = ListMCRResourceTags(&cobra.Command{Use: "list-tags"}, []string{mcrUID}, true, "json")
	})
	require.NoError(t, verifyTagsErr, "list MCR tags after clear output: %s", verifyTagsOut)
	cleared := tagsFromListJSON(t, verifyTagsOut)
	for k := range want {
		assert.NotContainsf(t, cleared, k, "tag %q should be cleared", k)
	}

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
	var createErr error
	createOut := captureTableOutput(func() {
		createErr = CreateMCRPrefixFilterList(createCmd, []string{mcrUID}, true)
	})
	require.NoError(t, createErr, "create prefix filter list output: %s", createOut)

	pflID := parsePrefixFilterListID(createOut)
	require.NotEmpty(t, pflID, "could not parse prefix filter list ID from: %s", createOut)

	// Best-effort cleanup. Registered after the MCR cleanup so it runs first
	// (cleanups run LIFO) — the list must be gone before the MCR is deleted.
	t.Cleanup(func() {
		var delErr error
		out := captureTableOutput(func() {
			delErr = DeleteMCRPrefixFilterList(&cobra.Command{Use: "delete"}, []string{mcrUID, pflID}, true)
		})
		if delErr != nil {
			t.Logf("cleanup: delete prefix filter list %s failed: %v; output: %s", pflID, delErr, out)
			return
		}
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
	updatePFLOut := captureTableOutput(func() {
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
	var deletePFLErr error
	deletePFLOut := captureTableOutput(func() {
		deletePFLErr = DeleteMCRPrefixFilterList(&cobra.Command{Use: "delete"}, []string{mcrUID, pflID}, true)
	})
	require.NoError(t, deletePFLErr, "delete prefix filter list output: %s", deletePFLOut)

	goneErr := GetMCRPrefixFilterList(&cobra.Command{Use: "get"}, []string{mcrUID, pflID}, true, "json")
	assert.Error(t, goneErr, "prefix filter list should no longer exist")
}

func TestIntegration_MCRJSONInputLifecycle(t *testing.T) {
	testutil.RequireStagingForProvisioning(t)
	client := testutil.SetupIntegrationClient(t)
	// Restore login via t.Cleanup, not defer: defers run before t.Cleanup, so a
	// deferred restore would swap back the default login (wrong environment)
	// before the resource-deletion cleanups below get to run.
	t.Cleanup(testutil.LoginWithClient(t, client))

	// Action functions mutate the process-wide output format; restore it so
	// test order can't leak state between tests in this package.
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

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

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR (JSON) output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut, "MCR")
	// Register cleanup before asserting on mcrUID, so any created MCR is
	// deleted even if the UID parse fails.
	t.Cleanup(func() {
		if mcrUID == "" {
			return
		}
		delCmd := integrationMCRDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMCR(delCmd, []string{mcrUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MCR %s failed: %v; output: %s", mcrUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MCR %s: %s", mcrUID, out)
	})
	require.NotEmpty(t, mcrUID, "could not parse MCR UID from: %s", buyOut)

	mcr := getMCRJSON(t, mcrUID)
	assert.Equal(t, mcrUID, mcr["uid"])
	assert.Equal(t, name, mcr["name"])

	newName := name + "-updated"
	updCmd := integrationMCRUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := captureTableOutput(func() { updErr = UpdateMCR(updCmd, []string{mcrUID}, true) })
	require.NoError(t, updErr, "update MCR output: %s", updOut)

	assert.Equal(t, newName, getMCRJSON(t, mcrUID)["name"])
}
