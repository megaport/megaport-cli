//go:build integration

package partners

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationListPartnersCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("product-name", "", "")
	cmd.Flags().String("connect-type", "", "")
	cmd.Flags().String("company-name", "", "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().Int("limit", 0, "")
	return cmd
}

func TestIntegration_ListPartners(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListPartnersCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = ListPartners(cmd, nil, true, "json")
	})

	require.NoError(t, err)
	require.NotEmpty(t, captured, "expected JSON output; ListPartners prints nothing for an empty result set in json mode")

	var partners []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &partners), "output should be valid JSON")
	assert.NotEmpty(t, partners, "staging should return at least one partner port")

	for _, p := range partners {
		assert.Contains(t, p, "product_name")
		assert.Contains(t, p, "uid")
		assert.Contains(t, p, "connect_type")
		assert.Contains(t, p, "company_name")
	}
}

func TestIntegration_ListPartners_FilterByConnectType(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListPartnersCmd()
	require.NoError(t, cmd.Flags().Set("connect-type", "AWS"))

	var err error
	captured := output.CaptureOutput(func() {
		err = ListPartners(cmd, nil, true, "json")
	})

	require.NoError(t, err)
	require.NotEmpty(t, captured, "expected JSON output; ListPartners prints nothing for an empty result set in json mode")

	var partners []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &partners))
	assert.NotEmpty(t, partners, "AWS partner ports should exist on staging")

	for _, p := range partners {
		ct, ok := p["connect_type"].(string)
		assert.Truef(t, ok, "connect_type should be a string in result %v", p)
		assert.Truef(t, strings.EqualFold("AWS", ct), "filtered results should all have connect_type AWS (case-insensitive), got %q", ct)
	}
}
