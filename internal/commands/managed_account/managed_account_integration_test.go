//go:build integration

package managed_account

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

func integrationListManagedAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Int("limit", 0, "")
	return cmd
}

func TestIntegration_ListManagedAccounts(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListManagedAccountsCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = ListManagedAccounts(cmd, nil, true, "json")
	})

	if err != nil {
		if strings.Contains(err.Error(), "not configured to create managed companies") ||
			strings.Contains(err.Error(), "403") {
			t.Skip("staging account not configured for managed companies — skipping")
		}
		require.NoError(t, err)
	}

	// An empty result with json format produces no output — that's acceptable for
	// staging accounts that have no managed accounts configured.
	if captured == "" {
		t.Skip("no managed accounts on staging — acceptable")
	}

	var accounts []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &accounts), "output should be valid JSON")

	for _, a := range accounts {
		assert.Contains(t, a, "account_name")
		assert.Contains(t, a, "company_uid")
	}
}

func TestIntegration_GetManagedAccount(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := integrationListManagedAccountsCmd()

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListManagedAccounts(listCmd, nil, true, "json")
	})
	if listErr != nil {
		if strings.Contains(listErr.Error(), "not configured to create managed companies") ||
			strings.Contains(listErr.Error(), "403") {
			t.Skip("staging account not configured for managed companies — skipping")
		}
		require.NoError(t, listErr)
	}

	if listOut == "" {
		t.Skip("no managed accounts on staging to test GetManagedAccount")
	}

	var accounts []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &accounts))
	require.NotEmpty(t, accounts, "expected non-empty accounts after non-empty JSON output")

	companyUID, ok := accounts[0]["company_uid"].(string)
	require.True(t, ok, "company_uid should be a string")
	accountName, ok := accounts[0]["account_name"].(string)
	require.True(t, ok, "account_name should be a string")

	getCmd := &cobra.Command{Use: "get"}

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetManagedAccount(getCmd, []string{companyUID, accountName}, true, "json")
	})

	require.NoError(t, getErr)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	assert.Equal(t, companyUID, got[0]["company_uid"])
	assert.Equal(t, accountName, got[0]["account_name"])
}
