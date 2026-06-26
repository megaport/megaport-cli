//go:build integration && provisioning

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
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// natGatewayUIDByName resolves a NAT gateway's UID by its unique product name.
// CreateNATGateway prints the UID but does not return it, so the lifecycle test
// reads it back from the SDK rather than scraping the success-line output.
func natGatewayUIDByName(client *megaport.Client, name string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	gws, err := client.NATGatewayService.ListNATGateways(ctx)
	if err != nil {
		return "", false, err
	}
	for _, gw := range gws {
		if gw.ProductName == name {
			return gw.ProductUID, true, nil
		}
	}
	return "", false, nil
}

func TestIntegration_NATGatewayLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	// Discover a valid speed tier from the session options.
	sessCmd := newTestCmd("list-sessions")

	var sessErr error
	sessOut := output.CaptureStdout(func() {
		sessErr = ListNATGatewaySessions(sessCmd, nil, true, "json")
	})
	require.NoError(t, sessErr)

	// This test is provisioning-tagged and only runs in the manual provisioning
	// job, where NAT Gateway is expected to be available. Assert availability
	// rather than skipping so a regression that removes it fails loudly instead
	// of going green having exercised nothing.
	require.NotEmpty(t, sessOut, "NAT gateway session options must be available in the provisioning environment")
	var sessions []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(sessOut), &sessions))
	require.NotEmpty(t, sessions, "NAT gateway session options must be available in the provisioning environment")

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
	require.NotEmpty(t, eligible, "NAT Gateway must be available at a provisioning-environment location for %d Mbps", speedMbps)
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

	// Register cleanup before create so the gateway is removed even if a later
	// assertion fails. If the UID was never captured (create succeeded but a
	// later step failed first), resolve it by the unique name via the SDK so
	// the gateway is not leaked.
	var createdUID string
	t.Cleanup(func() {
		uid := createdUID
		if uid == "" {
			if found, ok, err := natGatewayUIDByName(client, testName); err == nil && ok {
				uid = found
			}
		}
		if uid == "" {
			return
		}
		delCmd := newTestCmd("delete")
		_ = delCmd.Flags().Set("force", "true")
		if err := DeleteNATGateway(delCmd, []string{uid}, true); err != nil {
			t.Errorf("cleanup: failed to delete NAT gateway %s: %v", uid, err)
		}
	})

	var createErr error
	output.CaptureOutput(func() {
		createErr = CreateNATGateway(createCmd, nil, true)
	})
	require.NoError(t, createErr, "CreateNATGateway failed at staging location %d (speed %d Mbps)", locationID, speedMbps)

	// Resolve the UID from the SDK by the unique name rather than scraping the
	// create command's success line, so a future output-format change cannot
	// silently break UID capture.
	uid, ok, err := natGatewayUIDByName(client, testName)
	require.NoError(t, err, "SDK ListNATGateways failed after create")
	require.True(t, ok, "created NAT gateway %q not found via SDK", testName)
	createdUID = uid

	// Verify the gateway can be retrieved.
	getCmd := newTestCmd("get")

	var getErr error
	getOut := output.CaptureStdout(func() {
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
	// Regression guard for the flag-path fix: --asn must round-trip onto the
	// created gateway. JSON numbers decode as float64.
	assert.Equal(t, float64(64512), gw["asn"], "created gateway should retain the --asn flag value")

	// Validate the gateway to confirm pricing data is accessible.
	valCmd := newTestCmd("validate")
	var valErr error
	valOut := output.CaptureStdout(func() {
		valErr = ValidateNATGateway(valCmd, []string{uid}, true, "json")
	})
	require.NoError(t, valErr)
	var valItems []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(valOut), &valItems), "validate output should be valid JSON: %q", valOut)
	require.Len(t, valItems, 1)
	assert.Equal(t, uid, valItems[0]["uid"])
	assert.Contains(t, valItems[0], "monthly_rate")

	// Buy (provision) the gateway. This is the step that starts billing.
	buyCmd := newTestCmd("buy")
	require.NoError(t, buyCmd.Flags().Set("output", "json"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))
	var buyErr error
	// Capture stdout only: buy prints a "✓ NAT Gateway created" success line to
	// stderr, which would otherwise corrupt the JSON parsed from buyOut below.
	buyOut := output.CaptureStdout(func() {
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
	getOut2 := output.CaptureStdout(func() {
		getErr2 = GetNATGateway(getCmd2, []string{uid}, true, "json")
	})
	require.NoError(t, getErr2)

	var items2 []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut2), &items2))
	require.Len(t, items2, 1)
	assert.Equal(t, updatedName, items2[0]["name"])

	// Telemetry: assert no error and valid JSON; data may be empty for a new gateway.
	telCmd := newTestCmd("telemetry")
	require.NoError(t, telCmd.Flags().Set("types", "BITS"))
	require.NoError(t, telCmd.Flags().Set("days", "1"))

	var telErr error
	telOut := output.CaptureStdout(func() {
		telErr = GetNATGatewayTelemetry(telCmd, []string{uid}, true, "json")
	})
	require.NoError(t, telErr)

	// Telemetry serialises as a JSON array of samples; assert that shape rather
	// than just "valid JSON" so a malformed (non-array) response is caught. The
	// array may be empty for a freshly provisioned gateway.
	var telSamples []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(telOut), &telSamples), "telemetry output should be a JSON array of samples: %q", telOut)
}
