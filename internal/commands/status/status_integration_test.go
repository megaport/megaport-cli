//go:build integration

package status

import (
	"encoding/json"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationStatusCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "status"}
	cmd.Flags().Bool("include-inactive", false, "")
	return cmd
}

func TestIntegration_StatusDashboard(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationStatusCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = StatusDashboard(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var dashboard map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &dashboard), "output should be valid JSON")

	// The dashboard always renders all resource buckets plus a summary,
	// regardless of whether the account holds any resources.
	assert.Contains(t, dashboard, "ports")
	assert.Contains(t, dashboard, "mcrs")
	assert.Contains(t, dashboard, "mves")
	assert.Contains(t, dashboard, "vxcs")
	assert.Contains(t, dashboard, "ixs")
	assert.Contains(t, dashboard, "summary")

	summary, ok := dashboard["summary"].(map[string]interface{})
	require.True(t, ok, "summary should be an object")
	assert.Contains(t, summary, "ports")
	assert.Contains(t, summary, "vxcs")
}
