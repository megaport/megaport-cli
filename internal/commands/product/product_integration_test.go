//go:build integration

package product

import (
	"encoding/json"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationListProductsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	return cmd
}

func TestIntegration_ListProducts(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListProductsCmd()
	require.NoError(t, cmd.Flags().Set("include-inactive", "true"))

	var err error
	captured := output.CaptureOutput(func() {
		err = ListProducts(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var products []map[string]interface{}
	if captured != "" {
		require.NoError(t, json.Unmarshal([]byte(captured), &products), "output should be valid JSON")
	}
	if len(products) == 0 {
		t.Skip("staging account has no products provisioned")
	}

	for _, p := range products {
		assert.Contains(t, p, "uid", "product should have a uid field")
		assert.Contains(t, p, "type", "product should have a type field")
	}
}

func TestIntegration_GetProductType(t *testing.T) {
	// List first to get a valid product UID, then resolve its type.
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := integrationListProductsCmd()
	require.NoError(t, listCmd.Flags().Set("include-inactive", "true"))

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListProducts(listCmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	var products []map[string]interface{}
	if listOut != "" {
		require.NoError(t, json.Unmarshal([]byte(listOut), &products))
	}
	if len(products) == 0 {
		t.Skip("staging account has no products to resolve a type for")
	}

	uid, ok := products[0]["uid"].(string)
	require.True(t, ok, "product uid should be a string")
	require.NotEmpty(t, uid)

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetProductType(&cobra.Command{Use: "get-type"}, []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var types []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &types), "output should be valid JSON")
	require.NotEmpty(t, types, "get-type should return a result for a valid UID")
	assert.Contains(t, types[0], "uid", "result should have a uid field")
	assert.Contains(t, types[0], "type", "result should have a type field")
}
