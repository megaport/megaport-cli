//go:build integration

package servicekeys

import (
	"encoding/json"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationListServiceKeysCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("limit", 0, "")
	return cmd
}

func TestIntegration_ListServiceKeys(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := integrationListServiceKeysCmd()

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListServiceKeys(listCmd, nil, true, "json")
	})

	require.NoError(t, listErr)

	if listOut == "" {
		t.Skip("no service keys on staging")
	}

	var keys []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &keys), "list output should be valid JSON")

	if len(keys) == 0 {
		t.Skip("no service keys on staging")
	}

	for _, k := range keys {
		assert.Contains(t, k, "key_uid")
		assert.Contains(t, k, "product_uid")
		assert.Contains(t, k, "description")
	}
}

func TestIntegration_GetServiceKey(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := integrationListServiceKeysCmd()

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListServiceKeys(listCmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	if listOut == "" {
		t.Skip("no service keys on staging to test GetServiceKey")
	}

	var keys []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &keys), "list output should be valid JSON")
	require.NotEmpty(t, keys, "need at least one service key to test GetServiceKey")

	keyUID, ok := keys[0]["key_uid"].(string)
	require.True(t, ok, "key_uid should be a string")
	require.NotEmpty(t, keyUID)

	getCmd := &cobra.Command{Use: "get"}

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetServiceKey(getCmd, []string{keyUID}, true, "json")
	})

	require.NoError(t, getErr)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	assert.Equal(t, keyUID, got[0]["key_uid"])
	assert.Contains(t, got[0], "product_uid")
}
