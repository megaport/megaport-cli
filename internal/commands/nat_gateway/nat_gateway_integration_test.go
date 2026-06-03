//go:build integration

package nat_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_NATGatewayListSessions(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := newTestCmd("list-sessions")

	var err error
	captured := output.CaptureOutput(func() {
		err = ListNATGatewaySessions(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var sessions []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &sessions), "output should be valid JSON")
	if len(sessions) == 0 {
		t.Skip("no NAT gateway session options available on staging")
	}

	for _, s := range sessions {
		assert.Contains(t, s, "speed_mbps", "session entry should have speed_mbps field")
		assert.Contains(t, s, "session_counts", "session entry should have session_counts field")
	}
}

func TestIntegration_NATGatewayListAndGet(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := newTestCmd("list")

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListNATGateways(listCmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	var gateways []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &gateways))
	if len(gateways) == 0 {
		t.Skip("no NAT gateways on staging to test Get")
	}

	uid, ok := gateways[0]["uid"].(string)
	require.True(t, ok, "first gateway should have a uid string field")

	getCmd := newTestCmd("get")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetNATGateway(getCmd, []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &items))
	require.Len(t, items, 1)

	gw := items[0]
	assert.Equal(t, uid, gw["uid"])
	assert.Contains(t, gw, "name")
	assert.Contains(t, gw, "provisioning_status")
}

func TestIntegration_NATGatewayLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	// Discover a valid speed tier from the session options.
	sessCmd := newTestCmd("list-sessions")

	var sessErr error
	sessOut := output.CaptureOutput(func() {
		sessErr = ListNATGatewaySessions(sessCmd, nil, true, "json")
	})
	require.NoError(t, sessErr)

	var sessions []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(sessOut), &sessions))
	if len(sessions) == 0 {
		t.Skip("no NAT gateway session options available on staging")
	}

	rawSpeed, ok := sessions[0]["speed_mbps"].(float64)
	require.True(t, ok, "speed_mbps should be a number")
	speedMbps := int(rawSpeed)

	// Find a location that supports NAT Gateway at this speed.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	locs, err := client.LocationService.ListLocationsV3(ctx)
	require.NoError(t, err)
	eligible := client.LocationService.FilterLocationsByNATGatewaySpeedV3(ctx, speedMbps, locs)
	if len(eligible) == 0 {
		t.Skipf("NAT Gateway not available at any staging location for %d Mbps", speedMbps)
	}
	locationID := eligible[0].ID

	// Create a uniquely-named NAT gateway.
	testName := fmt.Sprintf("CLI-Test-NAT-%d", time.Now().UnixNano())

	createCmd := newTestCmd("create")
	require.NoError(t, createCmd.Flags().Set("name", testName))
	require.NoError(t, createCmd.Flags().Set("term", "1"))
	require.NoError(t, createCmd.Flags().Set("speed", fmt.Sprintf("%d", speedMbps)))
	require.NoError(t, createCmd.Flags().Set("location-id", fmt.Sprintf("%d", locationID)))
	require.NoError(t, createCmd.Flags().Set("yes", "true"))

	// Use table format so the "created <uid>" success line goes to stdout.
	// Restore whatever format was active before so later CaptureOutput calls are unaffected.
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })
	output.SetOutputFormat("table")

	var createErr error
	output.CaptureOutput(func() {
		createErr = CreateNATGateway(createCmd, nil, true)
	})
	if createErr != nil {
		t.Skipf("NAT Gateway not available at this staging location (ID %d): %v", locationID, createErr)
	}

	// Discover the UID by listing with the unique test name.
	findCmd := newTestCmd("list")
	require.NoError(t, findCmd.Flags().Set("name", testName))
	require.NoError(t, findCmd.Flags().Set("include-inactive", "true"))

	var findErr error
	findOut := output.CaptureOutput(func() {
		findErr = ListNATGateways(findCmd, nil, true, "json")
	})
	require.NoError(t, findErr)

	var found []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(findOut), &found))
	require.NotEmpty(t, found, "should find the newly created NAT gateway by name")

	uid, ok := found[0]["uid"].(string)
	require.True(t, ok, "created gateway should have a uid field")

	// Register cleanup before any assertions that might abort the test.
	t.Cleanup(func() {
		delCmd := newTestCmd("delete")
		_ = delCmd.Flags().Set("force", "true")
		if err := DeleteNATGateway(delCmd, []string{uid}, true); err != nil {
			t.Logf("cleanup: failed to delete NAT gateway %s: %v", uid, err)
		}
	})

	// Verify the gateway can be retrieved.
	getCmd := newTestCmd("get")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetNATGateway(getCmd, []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &items))
	require.Len(t, items, 1)
	gw := items[0]
	assert.Equal(t, uid, gw["uid"])
	assert.Equal(t, testName, gw["name"])
	assert.Contains(t, gw, "provisioning_status")

	// Update the name.
	updatedName := testName + "-upd"
	updateCmd := newTestCmd("update")
	require.NoError(t, updateCmd.Flags().Set("name", updatedName))

	var updateErr error
	output.CaptureOutput(func() {
		updateErr = UpdateNATGateway(updateCmd, []string{uid}, true)
	})
	require.NoError(t, updateErr)

	// Confirm the name change was applied.
	getCmd2 := newTestCmd("get")

	var getErr2 error
	getOut2 := output.CaptureOutput(func() {
		getErr2 = GetNATGateway(getCmd2, []string{uid}, true, "json")
	})
	require.NoError(t, getErr2)

	var items2 []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut2), &items2))
	require.Len(t, items2, 1)
	assert.Equal(t, updatedName, items2[0]["name"])

	// Telemetry — assert no error and valid JSON; data may be empty for a new gateway.
	telCmd := newTestCmd("telemetry")
	require.NoError(t, telCmd.Flags().Set("types", "BITS"))
	require.NoError(t, telCmd.Flags().Set("days", "1"))

	var telErr error
	telOut := output.CaptureOutput(func() {
		telErr = GetNATGatewayTelemetry(telCmd, []string{uid}, true, "json")
	})
	require.NoError(t, telErr)

	var telData interface{}
	require.NoError(t, json.Unmarshal([]byte(telOut), &telData), "telemetry output should be valid JSON: %q", telOut)
}
