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
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stagingMVELocationID is tried first when probing for MVE capacity. It is not
// authoritative: when it is out of capacity, findMVECapacity sweeps every other
// active MVE-capable location before giving up (mirroring the terraform
// provider's acceptance tests).
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

// productSizeCandidates returns the image's supported sizes ordered for a
// capacity probe: MEDIUM first (the size we want to exercise), then SMALL and
// the larger sizes, then any other size the image advertises. A given location
// can be out of capacity for one size but have room for another, so the probe
// tries each in turn. Label-format strings (e.g. "MVE 4/16") are normalized to
// programmatic names ("MEDIUM").
func (d discoveredImage) productSizeCandidates() []string {
	supported := make(map[string]bool, len(d.AvailableSizes))
	for _, s := range d.AvailableSizes {
		canonical := validation.NormalizeMVEProductSize(strings.ToUpper(s))
		// Skip advertised sizes we can't map to a programmatic name the buy API
		// accepts. Probing a raw label like "MVE 16/64" returns an invalid-size
		// error, not a capacity error, which would fail the probe.
		if validation.ValidateMVEProductSize(canonical) == nil {
			supported[canonical] = true
		}
	}
	var ordered []string
	emitted := make(map[string]bool, len(supported))
	add := func(size string) {
		if supported[size] && !emitted[size] {
			ordered = append(ordered, size)
			emitted[size] = true
		}
	}
	for _, size := range []string{"MEDIUM", "SMALL", "LARGE", "X_LARGE_12"} {
		add(size)
	}
	// Probe any other supported size after the preferred order, so an image
	// whose sizes fall outside the list above is still tried, not skipped.
	for _, s := range d.AvailableSizes {
		add(validation.NormalizeMVEProductSize(strings.ToUpper(s)))
	}
	if len(ordered) == 0 {
		ordered = append(ordered, "MEDIUM")
	}
	return ordered
}

// isMVECapacityError reports whether a buy/validate error is staging telling us
// it has no host capacity for the request, rather than a genuine failure. Such a
// shortage is environmental, so tests skip on it instead of failing.
func isMVECapacityError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "no available capacity")
}

// findMVECapacity probes staging for a location and product size with available
// Aruba MVE capacity, mirroring the terraform provider's acceptance tests: it
// validates the real order against the preferred location first, then sweeps
// every active MVE-capable location, returning the first (location, size) the
// API accepts. It skips the test only when no location in staging has capacity
// for any supported size, so a shortage at one site no longer fails the build.
func findMVECapacity(t *testing.T, client *megaport.Client, img discoveredImage) (locationID int, size string) {
	t.Helper()
	listCtx, listCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer listCancel()

	locations, err := client.LocationService.ListLocationsV3(listCtx)
	require.NoError(t, err, "list locations for MVE capacity probe")

	// Each probe gets its own deadline rather than sharing one across the sweep:
	// a shared budget could be drained by a few slow calls and skip later
	// locations that actually have capacity.
	probe := func(locID int, size string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := client.MVEService.ValidateMVEOrder(ctx, &megaport.BuyMVERequest{
			LocationID: locID,
			Name:       "cli-capacity-probe",
			Term:       1,
			VendorConfig: &megaport.ArubaConfig{
				Vendor:      "aruba",
				ImageID:     img.ID,
				ProductSize: size,
				AccountName: "test",
				AccountKey:  "test",
				SystemTag:   "test",
			},
			Vnics: []megaport.MVENetworkInterface{
				{Description: "MVE VNIC 1", VLAN: 55},
				{Description: "MVE VNIC 2", VLAN: 56},
			},
		})
		if err == nil {
			return true
		}
		require.True(t, isMVECapacityError(err), "validate MVE order at location %d size %s", locID, size)
		return false
	}

	// Probe the preferred location first, then every other active MVE-capable
	// location.
	ordered := make([]*megaport.LocationV3, 0, len(locations))
	for _, loc := range locations {
		if loc == nil {
			continue
		}
		if loc.ID == stagingMVELocationID && loc.IsStatusOrderable() && loc.HasMVESupport() {
			ordered = append(ordered, loc)
		}
	}
	for _, loc := range locations {
		if loc == nil || loc.ID == stagingMVELocationID {
			continue
		}
		if loc.IsStatusOrderable() && loc.HasMVESupport() {
			ordered = append(ordered, loc)
		}
	}

	sizes := img.productSizeCandidates()
	for _, loc := range ordered {
		for _, size := range sizes {
			if probe(loc.ID, size) {
				t.Logf("findMVECapacity: using location %d (%s) size %s", loc.ID, loc.Name, size)
				return loc.ID, size
			}
		}
	}
	t.Skip("no staging location has available Aruba MVE capacity for any supported size")
	return 0, ""
}

// buyMVEAtAvailableLocation discovers a location+size with capacity, runs BuyMVE
// once with the command built by buildCmd, and returns the created MVE's UID with
// a hard-delete cleanup registered. If capacity vanishes between the probe and
// the buy (a race against other consumers of shared staging), it skips rather
// than failing.
func buyMVEAtAvailableLocation(t *testing.T, client *megaport.Client, img discoveredImage, buildCmd func(locationID int, size string) *cobra.Command) string {
	t.Helper()
	locationID, size := findMVECapacity(t, client, img)

	cmd := buildCmd(locationID, size)
	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMVE(cmd, nil, true) })
	if isMVECapacityError(buyErr) {
		t.Skipf("MVE capacity at location %d (size %s) disappeared between probe and buy", locationID, size)
	}
	require.NoError(t, buyErr, "buy MVE output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	// Register cleanup before asserting on the UID, so any created MVE is torn
	// down even if the parse below fails.
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
	return mveUID
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
	vnics := `[{"description": "MVE VNIC 1", "vlan": 55}, {"description": "MVE VNIC 2", "vlan": 56}]`

	// BuyMVE waits for provisioning (no --no-wait), so the MVE is ready once it
	// returns. The buy retries across product sizes if staging is out of
	// capacity, and skips the test if every size is exhausted.
	mveUID := buyMVEAtAvailableLocation(t, client, img, func(locationID int, size string) *cobra.Command {
		vendorConfig := fmt.Sprintf(`{
			"vendor": "aruba",
			"productSize": "%s",
			"imageId": %d,
			"accountName": "test",
			"accountKey": "test",
			"systemTag": "test"
		}`, size, img.ID)
		buyCmd := integrationMVEBuyCmd()
		require.NoError(t, buyCmd.Flags().Set("name", name))
		require.NoError(t, buyCmd.Flags().Set("term", "1"))
		require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(locationID)))
		require.NoError(t, buyCmd.Flags().Set("vendor-config", vendorConfig))
		require.NoError(t, buyCmd.Flags().Set("vnics", vnics))
		require.NoError(t, buyCmd.Flags().Set("yes", "true"))
		return buyCmd
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

	mveUID := buyMVEAtAvailableLocation(t, client, img, func(locationID int, size string) *cobra.Command {
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
		}`, name, locationID, size, img.ID)
		buyCmd := integrationMVEBuyCmd()
		require.NoError(t, buyCmd.Flags().Set("json", buyJSON))
		return buyCmd
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
	updOut := captureTableOutput(func() { updErr = UpdateMVE(updCmd, []string{mveUID}, true) })
	require.NoError(t, updErr, "update MVE output: %s", updOut)

	assert.Equal(t, newName, getMVEJSON(t, mveUID)["name"])
}
