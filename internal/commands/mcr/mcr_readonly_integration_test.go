//go:build integration

package mcr

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

func readOnlyListMCRsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Int("port-speed", 0, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().StringArray("tag", nil, "")
	return cmd
}

func readOnlyGetMCRCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("watch", false, "")
	cmd.Flags().Bool("export", false, "")
	return cmd
}

func readOnlyStatusMCRCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "status"}
	cmd.Flags().Bool("watch", false, "")
	return cmd
}

// TestIntegration_MCRReadOnly is a fast read-only smoke test against the
// configured environment (staging by default):
// list, then get + status on the first MCR. Skips cleanly when the account has
// no MCRs. Performs no mutation.
func TestIntegration_MCRReadOnly(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)
	origFormat := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFormat) })

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListMCRs(readOnlyListMCRsCmd(), nil, true, "json")
	})
	require.NoError(t, listErr)

	if strings.TrimSpace(listOut) == "" {
		t.Skip("no MCRs on the account")
	}
	var mcrList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &mcrList), "ListMCRs returned invalid JSON")
	if len(mcrList) == 0 {
		t.Skip("no MCRs on the account")
	}

	first := mcrList[0]
	assert.Contains(t, first, "uid", "MCR should have a uid field")
	assert.Contains(t, first, "name", "MCR should have a name field")

	uid, ok := first["uid"].(string)
	require.True(t, ok && uid != "", "uid must be a non-empty string")

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetMCR(readOnlyGetMCRCmd(), []string{uid}, true, "json")
	})
	require.NoError(t, getErr)

	var gotList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &gotList), "GetMCR JSON output must be valid")
	require.Len(t, gotList, 1)
	assert.Equal(t, uid, gotList[0]["uid"])
	assert.Contains(t, gotList[0], "name")

	var statusErr error
	statusOut := output.CaptureOutput(func() {
		statusErr = GetMCRStatus(readOnlyStatusMCRCmd(), []string{uid}, true, "json")
	})
	require.NoError(t, statusErr)

	var statusList []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(statusOut), &statusList), "GetMCRStatus JSON output must be valid")
	require.Len(t, statusList, 1)
	assert.Equal(t, uid, statusList[0]["uid"])
	assert.Contains(t, statusList[0], "status")
}
