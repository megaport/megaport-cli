//go:build integration

package nat_gateway

import (
	"encoding/json"
	"testing"

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
