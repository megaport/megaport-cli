//go:build integration

package mve

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/validation"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stagingMVELocationID is a known MVE-capable staging location. Discovering one
// via ListLocations would be overengineering for this lifecycle test.
const stagingMVELocationID = 65

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

func integrationMVEBuyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-wait", false, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("vendor-config", "", "")
	cmd.Flags().String("vnics", "", "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationMVEUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().String("vnics", "", "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func integrationMVEDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	return cmd
}

func integrationMVETagCmd() *cobra.Command {
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

type discoveredImage struct {
	ID             int      `json:"id"`
	Vendor         string   `json:"vendor"`
	ProductCode    string   `json:"productCode"`
	AvailableSizes []string `json:"availableSizes"`
}

// discoverArubaImage lists staging MVE images and returns the first Aruba image,
// skipping the test if none are available. The vendor is fixed (Aruba's vendor
// config is simple to construct); only the image ID is discovered dynamically,
// since image IDs change on staging.
func discoverArubaImage(t *testing.T) discoveredImage {
	t.Helper()
	var err error
	out := output.CaptureOutput(func() {
		err = ListMVEImages(&cobra.Command{Use: "list-images"}, nil, true, "json")
	})
	require.NoError(t, err, "list MVE images output: %s", out)

	var images []discoveredImage
	require.NoError(t, json.Unmarshal([]byte(out), &images), "images output should be valid JSON: %s", out)

	for _, img := range images {
		if strings.EqualFold(img.Vendor, "aruba") && img.ID > 0 {
			return img
		}
	}
	t.Skip("no Aruba MVE image available on staging")
	return discoveredImage{}
}

// productSize picks a size the image supports, preferring MEDIUM.
// It normalizes label-format strings (e.g. "MVE 4/16") to their programmatic
// equivalents (e.g. "MEDIUM") so the value is valid for the buy command.
func (d discoveredImage) productSize() string {
	for _, s := range d.AvailableSizes {
		normalized := validation.NormalizeMVEProductSize(strings.ToUpper(s))
		if normalized == "MEDIUM" {
			return "MEDIUM"
		}
	}
	if len(d.AvailableSizes) > 0 {
		return validation.NormalizeMVEProductSize(strings.ToUpper(d.AvailableSizes[0]))
	}
	return "MEDIUM"
}

// getMVEJSON retrieves an MVE as JSON and returns the decoded object.
func getMVEJSON(t *testing.T, uid string) map[string]interface{} {
	t.Helper()
	var err error
	out := output.CaptureOutput(func() {
		err = GetMVE(&cobra.Command{Use: "get"}, []string{uid}, true, "json")
	})
	require.NoError(t, err, "get MVE output: %s", out)

	var mves []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &mves), "MVE output should be valid JSON: %s", out)
	require.Len(t, mves, 1, "expected exactly one MVE")
	return mves[0]
}

func TestIntegration_MVELifecycle(t *testing.T) {
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

	img := discoverArubaImage(t)
	name := fmt.Sprintf("CLI-Test-MVE-%s", generateUniqueID())

	vendorConfig := fmt.Sprintf(`{
		"vendor": "aruba",
		"productSize": "%s",
		"imageId": %d,
		"accountName": "test",
		"accountKey": "test",
		"systemTag": "test"
	}`, img.productSize(), img.ID)
	vnics := `[{"description": "MVE VNIC 1", "vlan": 55}, {"description": "MVE VNIC 2", "vlan": 56}]`

	// BuyMVE waits for provisioning (no --no-wait), so the MVE is ready once it
	// returns.
	buyCmd := integrationMVEBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(stagingMVELocationID)))
	require.NoError(t, buyCmd.Flags().Set("vendor-config", vendorConfig))
	require.NoError(t, buyCmd.Flags().Set("vnics", vnics))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMVE(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MVE output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	// Register cleanup before asserting on mveUID, so any created MVE is
	// cleaned up even if the UID assertion below fails.
	t.Cleanup(func() {
		if mveUID == "" {
			return
		}
		delCmd := integrationMVEDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMVE(delCmd, []string{mveUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MVE %s failed: %v; output: %s", mveUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MVE %s: %s", mveUID, out)
	})
	require.NotEmpty(t, mveUID, "could not parse MVE UID from: %s", buyOut)

	// Get and verify the core fields.
	mve := getMVEJSON(t, mveUID)
	assert.Equal(t, mveUID, mve["uid"])
	assert.Equal(t, name, mve["name"])
	assert.Contains(t, mve, "status")
	assert.NotEmpty(t, mve["status"])
	assert.Equal(t, "aruba", strings.ToLower(fmt.Sprintf("%v", mve["vendor"])))

	// Update the name and verify it took effect.
	newName := name + "-updated"
	updCmd := integrationMVEUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := captureTableOutput(func() { updErr = UpdateMVE(updCmd, []string{mveUID}, true) })
	require.NoError(t, updErr, "update MVE output: %s", updOut)

	assert.Equal(t, newName, getMVEJSON(t, mveUID)["name"])

	// Resource tag round-trip (ESD-1392): set tags via update-tags, read them
	// back via list-tags, then clear them. Rides on the lifecycle MVE, so no
	// extra cleanup is needed.
	want := map[string]string{"env": "cli-integration", "owner": "esd-1392"}
	setTagsJSON, err := json.Marshal(want)
	require.NoError(t, err)
	setTagsCmd := integrationMVETagCmd()
	require.NoError(t, setTagsCmd.Flags().Set("json", string(setTagsJSON)))
	require.NoError(t, setTagsCmd.Flags().Set("force", "true"))
	var setTagsErr error
	setTagsOut := captureTableOutput(func() { setTagsErr = UpdateMVEResourceTags(setTagsCmd, []string{mveUID}, true) })
	require.NoError(t, setTagsErr, "update MVE tags output: %s", setTagsOut)

	var listTagsErr error
	listTagsOut := output.CaptureOutput(func() {
		listTagsErr = ListMVEResourceTags(&cobra.Command{Use: "list-tags"}, []string{mveUID}, true, "json")
	})
	require.NoError(t, listTagsErr, "list MVE tags output: %s", listTagsOut)
	// Assert our tags round-tripped without requiring the map to contain only
	// them, so an API-injected tag can't make this flaky.
	got := tagsFromListJSON(t, listTagsOut)
	for k, v := range want {
		assert.Equalf(t, v, got[k], "tag %q should round-trip", k)
	}

	// Clear the tags so the MVE is left clean for the steps that follow.
	clearTagsCmd := integrationMVETagCmd()
	require.NoError(t, clearTagsCmd.Flags().Set("json", "{}"))
	require.NoError(t, clearTagsCmd.Flags().Set("force", "true"))
	var clearTagsErr error
	clearTagsOut := captureTableOutput(func() { clearTagsErr = UpdateMVEResourceTags(clearTagsCmd, []string{mveUID}, true) })
	require.NoError(t, clearTagsErr, "clear MVE tags output: %s", clearTagsOut)

	var verifyTagsErr error
	verifyTagsOut := output.CaptureOutput(func() {
		verifyTagsErr = ListMVEResourceTags(&cobra.Command{Use: "list-tags"}, []string{mveUID}, true, "json")
	})
	require.NoError(t, verifyTagsErr, "list MVE tags after clear output: %s", verifyTagsOut)
	cleared := tagsFromListJSON(t, verifyTagsOut)
	for k := range want {
		assert.NotContainsf(t, cleared, k, "tag %q should be cleared", k)
	}

	// Update the vNIC descriptions via --vnics and verify they took effect.
	// The count and VLANs are immutable, so the update array must have one
	// entry per existing vNIC, applied in order.
	vnicUpdate := `[{"description": "MVE VNIC 1 updated"}, {"description": "MVE VNIC 2 updated"}]`
	vnicCmd := integrationMVEUpdateCmd()
	require.NoError(t, vnicCmd.Flags().Set("vnics", vnicUpdate))
	var vnicErr error
	vnicOut := captureTableOutput(func() { vnicErr = UpdateMVE(vnicCmd, []string{mveUID}, true) })
	require.NoError(t, vnicErr, "update MVE vNICs output: %s", vnicOut)

	// The CLI's JSON output doesn't expose vNIC descriptions, so read them
	// back through the SDK client directly.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	refreshed, err := client.MVEService.GetMVE(ctx, mveUID)
	require.NoError(t, err, "get MVE via SDK after vNIC update")
	require.Len(t, refreshed.NetworkInterfaces, 2)
	assert.Equal(t, "MVE VNIC 1 updated", refreshed.NetworkInterfaces[0].Description)
	assert.Equal(t, "MVE VNIC 2 updated", refreshed.NetworkInterfaces[1].Description)
}

func TestIntegration_MVEJSONInputLifecycle(t *testing.T) {
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

	img := discoverArubaImage(t)
	name := fmt.Sprintf("CLI-JSON-MVE-%s", generateUniqueID())

	buyJSON := fmt.Sprintf(`{
		"name": "%s",
		"term": 1,
		"locationId": %d,
		"vendorConfig": {
			"vendor": "aruba",
			"productSize": "%s",
			"imageId": %d,
			"accountName": "test",
			"accountKey": "test",
			"systemTag": "test"
		},
		"vnics": [{"description": "MVE VNIC 1", "vlan": 55}, {"description": "MVE VNIC 2", "vlan": 56}]
	}`, name, stagingMVELocationID, img.productSize(), img.ID)

	buyCmd := integrationMVEBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("json", buyJSON))

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMVE(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MVE (JSON) output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	// Register cleanup before asserting on mveUID, so any created MVE is
	// cleaned up even if the UID assertion below fails.
	t.Cleanup(func() {
		if mveUID == "" {
			return
		}
		delCmd := integrationMVEDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMVE(delCmd, []string{mveUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MVE %s failed: %v; output: %s", mveUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MVE %s: %s", mveUID, out)
	})
	require.NotEmpty(t, mveUID, "could not parse MVE UID from: %s", buyOut)

	mve := getMVEJSON(t, mveUID)
	assert.Equal(t, mveUID, mve["uid"])
	assert.Equal(t, name, mve["name"])
	assert.Contains(t, mve, "status")
	assert.Equal(t, "aruba", strings.ToLower(fmt.Sprintf("%v", mve["vendor"])))

	newName := name + "-updated"
	updCmd := integrationMVEUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := captureTableOutput(func() { updErr = UpdateMVE(updCmd, []string{mveUID}, true) })
	require.NoError(t, updErr, "update MVE output: %s", updOut)

	assert.Equal(t, newName, getMVEJSON(t, mveUID)["name"])
}
