//go:build integration

package mve

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
	cmd.Flags().Int("contract-term", 0, "")
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
func (d discoveredImage) productSize() string {
	for _, s := range d.AvailableSizes {
		if strings.EqualFold(s, "MEDIUM") {
			return "MEDIUM"
		}
	}
	if len(d.AvailableSizes) > 0 {
		return d.AvailableSizes[0]
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
	vnics := `[{"description": "MVE VNIC 1", "vlan": 55}]`

	// BuyMVE waits for provisioning (no --no-wait), so the MVE is ready once it
	// returns.
	buyCmd := integrationMVEBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(stagingMVELocationID)))
	require.NoError(t, buyCmd.Flags().Set("vendor-config", vendorConfig))
	require.NoError(t, buyCmd.Flags().Set("vnics", vnics))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	output.SetOutputFormat("table") // route the success message to stdout
	var buyErr error
	buyOut := output.CaptureOutput(func() { buyErr = BuyMVE(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MVE output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	require.NotEmpty(t, mveUID, "could not parse MVE UID from: %s", buyOut)

	t.Cleanup(func() {
		delCmd := integrationMVEDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		out := output.CaptureOutput(func() { _ = DeleteMVE(delCmd, []string{mveUID}, true) })
		t.Logf("cleanup: delete MVE %s: %s", mveUID, out)
	})

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
	updOut := output.CaptureOutput(func() { updErr = UpdateMVE(updCmd, []string{mveUID}, true) })
	require.NoError(t, updErr, "update MVE output: %s", updOut)

	assert.Equal(t, newName, getMVEJSON(t, mveUID)["name"])
}

func TestIntegration_MVEJSONInputLifecycle(t *testing.T) {
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
		"vnics": [{"description": "MVE VNIC 1", "vlan": 55}]
	}`, name, stagingMVELocationID, img.productSize(), img.ID)

	buyCmd := integrationMVEBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("json", buyJSON))

	output.SetOutputFormat("table")
	var buyErr error
	buyOut := output.CaptureOutput(func() { buyErr = BuyMVE(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MVE (JSON) output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	require.NotEmpty(t, mveUID, "could not parse MVE UID from: %s", buyOut)

	t.Cleanup(func() {
		delCmd := integrationMVEDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		out := output.CaptureOutput(func() { _ = DeleteMVE(delCmd, []string{mveUID}, true) })
		t.Logf("cleanup: delete MVE %s: %s", mveUID, out)
	})

	mve := getMVEJSON(t, mveUID)
	assert.Equal(t, mveUID, mve["uid"])
	assert.Equal(t, name, mve["name"])
	assert.Contains(t, mve, "status")
	assert.Equal(t, "aruba", strings.ToLower(fmt.Sprintf("%v", mve["vendor"])))

	newName := name + "-updated"
	updCmd := integrationMVEUpdateCmd()
	require.NoError(t, updCmd.Flags().Set("name", newName))
	var updErr error
	updOut := output.CaptureOutput(func() { updErr = UpdateMVE(updCmd, []string{mveUID}, true) })
	require.NoError(t, updErr, "update MVE output: %s", updOut)

	assert.Equal(t, newName, getMVEJSON(t, mveUID)["name"])
}
