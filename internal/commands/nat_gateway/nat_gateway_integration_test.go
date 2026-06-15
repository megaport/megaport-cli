//go:build integration

package nat_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_NATGatewayListSessionsReadOnly(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	cmd := newTestCmd("list-sessions")

	var err error
	captured := output.CaptureOutput(func() {
		err = ListNATGatewaySessions(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	if captured == "" {
		t.Skip("no NAT gateway session options available on staging")
	}
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

func TestIntegration_NATGatewayListAndGetReadOnly(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	listCmd := newTestCmd("list")

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListNATGateways(listCmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	if listOut == "" {
		t.Skip("no NAT gateways on staging to test Get")
	}
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
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })
	// Force normal verbosity so PrintResourceCreated always emits the UID to stdout;
	// quiet mode would suppress it and leave createdUID empty, leaking the gateway.
	origVerbosity := "normal"
	if output.IsQuiet() {
		origVerbosity = "quiet"
	} else if output.IsVerbose() {
		origVerbosity = "verbose"
	}
	output.SetVerbosity("normal")
	t.Cleanup(func() { output.SetVerbosity(origVerbosity) })

	// Discover a valid speed tier from the session options.
	sessCmd := newTestCmd("list-sessions")

	var sessErr error
	sessOut := output.CaptureOutput(func() {
		sessErr = ListNATGatewaySessions(sessCmd, nil, true, "json")
	})
	require.NoError(t, sessErr)

	if sessOut == "" {
		t.Skip("no NAT gateway session options available on staging")
	}
	var sessions []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(sessOut), &sessions))
	if len(sessions) == 0 {
		t.Skip("no NAT gateway session options available on staging")
	}

	rawSpeed, ok := sessions[0]["speed_mbps"].(float64)
	require.True(t, ok, "speed_mbps should be a number")
	speedMbps := int(rawSpeed)

	// Extract the first valid session count for the chosen speed tier.
	// The output formatter serialises SessionCount []int as a comma-separated
	// string (e.g. "1000, 2000"), so parse the first token.
	rawCounts, ok := sessions[0]["session_counts"].(string)
	require.True(t, ok, "session_counts should be a string")
	require.NotEmpty(t, rawCounts, "session_counts should not be empty")
	firstToken := strings.TrimSpace(strings.SplitN(rawCounts, ",", 2)[0])
	sessionCount, err := strconv.Atoi(firstToken)
	require.NoError(t, err, "failed to parse session count from %q", rawCounts)

	// Find a location that supports NAT Gateway at this speed.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	locs, err := client.LocationService.ListLocationsV3(ctx)
	require.NoError(t, err)
	eligible := client.LocationService.FilterLocationsByNATGatewaySpeedV3(ctx, speedMbps, locs)
	if len(eligible) == 0 {
		t.Skipf("NAT Gateway not available at any staging location for %d Mbps", speedMbps)
	}
	loc := eligible[0]
	locationID := loc.ID

	// Determine which diversity zone at this location supports the chosen speed.
	diversityZone := ""
	if loc.DiversityZones != nil {
		if loc.DiversityZones.Red != nil {
			for _, s := range loc.DiversityZones.Red.NATGatewaySpeedMbps {
				if s == speedMbps {
					diversityZone = "red"
					break
				}
			}
		}
		if diversityZone == "" && loc.DiversityZones.Blue != nil {
			for _, s := range loc.DiversityZones.Blue.NATGatewaySpeedMbps {
				if s == speedMbps {
					diversityZone = "blue"
					break
				}
			}
		}
	}
	require.NotEmpty(t, diversityZone, "no diversity zone at location %d supports speed %d Mbps", locationID, speedMbps)

	// Create a uniquely-named NAT gateway.
	testName := fmt.Sprintf("CLI-Test-NAT-%d", time.Now().UnixNano())

	createCmd := newTestCmd("create")
	require.NoError(t, createCmd.Flags().Set("name", testName))
	require.NoError(t, createCmd.Flags().Set("term", "1"))
	require.NoError(t, createCmd.Flags().Set("speed", fmt.Sprintf("%d", speedMbps)))
	require.NoError(t, createCmd.Flags().Set("location-id", fmt.Sprintf("%d", locationID)))
	require.NoError(t, createCmd.Flags().Set("session-count", fmt.Sprintf("%d", sessionCount)))
	require.NoError(t, createCmd.Flags().Set("asn", "64512"))
	require.NoError(t, createCmd.Flags().Set("diversity-zone", diversityZone))
	require.NoError(t, createCmd.Flags().Set("yes", "true"))

	// Use table format so CreateNATGateway's "✓ NAT Gateway created <uid>" line goes to stdout.
	output.SetOutputFormat("table")

	// Register cleanup before create so the gateway is deleted regardless of any
	// subsequent skip or assertion failure before createdUID is assigned below.
	var createdUID string
	t.Cleanup(func() {
		if createdUID == "" {
			return
		}
		delCmd := newTestCmd("delete")
		_ = delCmd.Flags().Set("force", "true")
		if err := DeleteNATGateway(delCmd, []string{createdUID}, true); err != nil {
			t.Errorf("cleanup: failed to delete NAT gateway %s: %v", createdUID, err)
		}
	})

	var createErr error
	createOut := output.CaptureOutput(func() {
		createErr = CreateNATGateway(createCmd, nil, true)
	})
	require.NoError(t, createErr, "CreateNATGateway failed at staging location %d (speed %d Mbps)", locationID, speedMbps)

	// Extract the UID from "✓ NAT Gateway created <uid>" and wire up cleanup immediately
	// so the gateway is deleted even if a subsequent skip fires.
	// Use Index rather than line-by-line CutPrefix: in TTY+table mode the spinner
	// writes \r-overwrite frames to stdout before the success line, so the prefix
	// may not appear at the start of a newline-split segment.
	const createPrefix = "✓ NAT Gateway created "
	if idx := strings.Index(createOut, createPrefix); idx >= 0 {
		after := createOut[idx+len(createPrefix):]
		if end := strings.IndexAny(after, "\r\n"); end >= 0 {
			createdUID = strings.TrimSpace(after[:end])
		} else {
			createdUID = strings.TrimSpace(after)
		}
	}
	require.NotEmpty(t, createdUID, "failed to parse NAT Gateway UID from create output: %q", createOut)
	uid := createdUID

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

	// Validate the gateway to confirm pricing data is accessible.
	valCmd := newTestCmd("validate")
	var valErr error
	valOut := output.CaptureOutput(func() {
		valErr = ValidateNATGateway(valCmd, []string{uid}, true, "json")
	})
	require.NoError(t, valErr)
	var valItems []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(valOut), &valItems), "validate output should be valid JSON: %q", valOut)
	require.Len(t, valItems, 1)
	assert.Equal(t, uid, valItems[0]["uid"])
	assert.Contains(t, valItems[0], "monthly_rate")

	// Buy (provision) the gateway — this is the step that starts billing.
	buyCmd := newTestCmd("buy")
	require.NoError(t, buyCmd.Flags().Set("output", "json"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))
	var buyErr error
	buyOut := output.CaptureOutput(func() {
		buyErr = BuyNATGateway(buyCmd, []string{uid}, true)
	})
	require.NoError(t, buyErr)
	var buyItems []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buyOut), &buyItems), "buy output should be valid JSON: %q", buyOut)
	require.Len(t, buyItems, 1)
	assert.Equal(t, uid, buyItems[0]["uid"])
	assert.Contains(t, buyItems[0], "provisioning_status")

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
