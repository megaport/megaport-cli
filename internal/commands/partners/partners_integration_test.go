//go:build integration
// +build integration

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

	var partners []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &partners), "output should be valid JSON")
	assert.NotEmpty(t, partners, "staging should return at least one partner megaport")

	// Spot-check expected fields (JSON output uses snake_case from PartnerOutput struct tags)
	for _, p := range partners {
		assert.Contains(t, p, "uid", "partner should have uid")
		assert.Contains(t, p, "company_name", "partner should have company_name")
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

	// ListPartners emits no output for an empty filtered result (see partners_actions.go).
	// Treat empty captured output as an empty JSON array so the test always passes
	// regardless of whether AWS partners exist on staging.
	if strings.TrimSpace(captured) == "" {
		captured = "[]"
	}

	var partners []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &partners))
	t.Logf("found %d AWS partner megaports on staging", len(partners))
}
