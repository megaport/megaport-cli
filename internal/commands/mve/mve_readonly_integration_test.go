//go:build integration

package mve

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

func readOnlyListMVEsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().String("vendor", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().StringArray("tag", nil, "")
	return cmd
}

func readOnlyGetMVECmd() *cobra.Command {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("watch", false, "")
	cmd.Flags().Bool("export", false, "")
	return cmd
}

func readOnlyStatusMVECmd() *cobra.Command {
	cmd := &cobra.Command{Use: "status"}
	cmd.Flags().Bool("watch", false, "")
	return cmd
}

// TestIntegration_MVEReadOnly is a fast read-only smoke test against staging:
// list, then get + status on the first MVE. Skips cleanly when the account has
// no MVEs. Performs no mutation.
func TestIntegration_MVEReadOnly(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	t.Cleanup(testutil.LoginWithClient(t, client))
	origFormat := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFormat) })

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListMVEs(readOnlyListMVEsCmd(), nil, true, "json")
	})
	require.NoError(t, listErr)

	if strings.TrimSpace(listOut) == "" {
		t.Skip("no MVEs on staging account")
	}
	var mveList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &mveList), "ListMVEs returned invalid JSON")
	if len(mveList) == 0 {
		t.Skip("no MVEs on staging account")
	}

	first := mveList[0]
	assert.Contains(t, first, "uid", "MVE should have a uid field")
	assert.Contains(t, first, "name", "MVE should have a name field")

	uid, ok := first["uid"].(string)
	require.True(t, ok && uid != "", "uid must be a non-empty string")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetMVE(readOnlyGetMVECmd(), []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var gotList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &gotList), "GetMVE JSON output must be valid")
	require.Len(t, gotList, 1)
	assert.Equal(t, uid, gotList[0]["uid"])
	assert.Contains(t, gotList[0], "name")

	var statusErr error
	statusOut := output.CaptureOutput(func() {
		statusErr = GetMVEStatus(readOnlyStatusMVECmd(), []string{uid}, true, "json")
	})
	require.NoError(t, statusErr)

	var statusList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(statusOut), &statusList), "GetMVEStatus JSON output must be valid")
	require.Len(t, statusList, 1)
	assert.Equal(t, uid, statusList[0]["uid"])
	assert.Contains(t, statusList[0], "status")
}
