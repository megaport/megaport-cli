//go:build integration
// +build integration

package locations

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationListLocationsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("metro", "", "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("market-code", "", "")
	cmd.Flags().Bool("mcr-available", false, "")
	return cmd
}

func TestIntegration_ListLocations(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListLocationsCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLocations(cmd, nil, true, "json")
	})

	require.NoError(t, err)

	var locs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &locs), "output should be valid JSON")
	assert.NotEmpty(t, locs, "staging should return at least one location")

	// Spot-check expected fields exist in each location
	for _, loc := range locs {
		assert.Contains(t, loc, "id", "location should have an id field")
		assert.Contains(t, loc, "name", "location should have a name field")
	}
}

func TestIntegration_ListLocations_FilterByCountry(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListLocationsCmd()
	require.NoError(t, cmd.Flags().Set("country", "Australia"))

	var err error
	captured := output.CaptureOutput(func() {
		err = ListLocations(cmd, nil, true, "json")
	})

	require.NoError(t, err)

	var locs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &locs))
	assert.NotEmpty(t, locs, "Australia should have at least one location on staging")
}

func TestIntegration_SearchLocations(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	var err error
	captured := output.CaptureOutput(func() {
		err = SearchLocations(&cobra.Command{Use: "search"}, []string{"Sydney"}, true, "json")
	})

	require.NoError(t, err)

	var locs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &locs))
	assert.NotEmpty(t, locs, "search for 'Sydney' should return results on staging")
}

func TestIntegration_GetLocation(t *testing.T) {
	// First list to get a valid ID, then retrieve individually
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListLocationsCmd()

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListLocations(cmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	var locs []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &locs))
	require.NotEmpty(t, locs, "need at least one location to test GetLocation")

	// Extract the first location's ID
	firstID, ok := locs[0]["id"].(float64)
	require.True(t, ok, "location id should be a number")
	locationID := int(firstID)

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetLocation(&cobra.Command{Use: "get"}, []string{formatLocationID(locationID)}, true, "json")
	})

	require.NoError(t, getErr)
	assert.NotEmpty(t, getOut)
}

// formatLocationID converts an int location ID to the string arg GetLocation expects.
func formatLocationID(id int) string {
	return fmt.Sprintf("%d", id)
}
