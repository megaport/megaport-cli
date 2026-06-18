//go:build integration

package vxc

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

func readOnlyListVXCsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("name-contains", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().String("a-end-uid", "", "")
	cmd.Flags().String("b-end-uid", "", "")
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().StringArray("tag", nil, "")
	return cmd
}

func readOnlyGetVXCCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("watch", false, "")
	cmd.Flags().Bool("export", false, "")
	return cmd
}

func readOnlyStatusVXCCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "status"}
	cmd.Flags().Bool("watch", false, "")
	return cmd
}

// TestIntegration_VXCReadOnly is a fast read-only smoke test against the
// configured environment (staging by default):
// list, then get + status on the first VXC. Skips cleanly when the account has
// no VXCs. Performs no mutation.
func TestIntegration_VXCReadOnly(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)
	origFormat := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFormat) })

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListVXCs(readOnlyListVXCsCmd(), nil, true, "json")
	})
	require.NoError(t, listErr)

	if strings.TrimSpace(listOut) == "" {
		t.Skip("no VXCs on the account")
	}
	var vxcList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &vxcList), "ListVXCs returned invalid JSON")
	if len(vxcList) == 0 {
		t.Skip("no VXCs on the account")
	}

	first := vxcList[0]
	assert.Contains(t, first, "uid", "VXC should have a uid field")
	assert.Contains(t, first, "name", "VXC should have a name field")

	uid, ok := first["uid"].(string)
	require.True(t, ok && uid != "", "uid must be a non-empty string")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetVXC(readOnlyGetVXCCmd(), []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var gotList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &gotList), "GetVXC JSON output must be valid")
	require.Len(t, gotList, 1)
	assert.Equal(t, uid, gotList[0]["uid"])
	assert.Contains(t, gotList[0], "name")

	var statusErr error
	statusOut := output.CaptureOutput(func() {
		statusErr = GetVXCStatus(readOnlyStatusVXCCmd(), []string{uid}, true, "json")
	})
	require.NoError(t, statusErr)

	var statusList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(statusOut), &statusList), "GetVXCStatus JSON output must be valid")
	require.Len(t, statusList, 1)
	assert.Equal(t, uid, statusList[0]["uid"])
	assert.Contains(t, statusList[0], "status")
}
