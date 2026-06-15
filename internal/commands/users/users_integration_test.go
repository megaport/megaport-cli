//go:build integration

package users

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationListUsersCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("position", "", "")
	cmd.Flags().Bool("active-only", false, "")
	cmd.Flags().Bool("inactive-only", false, "")
	return cmd
}

func TestIntegration_ListUsers(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationListUsersCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = ListUsers(cmd, nil, true, "json")
	})

	require.NoError(t, err)

	var users []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &users), "output should be valid JSON")
	assert.NotEmpty(t, users, "the authenticated account should have at least one user")

	for _, u := range users {
		assert.Contains(t, u, "employee_id")
		assert.Contains(t, u, "first_name")
		assert.Contains(t, u, "last_name")
		assert.Contains(t, u, "email")
	}
}

func TestIntegration_GetUser(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	listCmd := integrationListUsersCmd()

	var listErr error
	listOut := output.CaptureOutput(func() {
		listErr = ListUsers(listCmd, nil, true, "json")
	})
	require.NoError(t, listErr)

	var users []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(listOut), &users))
	require.NotEmpty(t, users, "need at least one user to test GetUser")

	empIDFloat, ok := users[0]["employee_id"].(float64)
	require.True(t, ok, "employee_id should be a number")
	employeeID := strconv.Itoa(int(empIDFloat))

	getCmd := &cobra.Command{Use: "get"}

	var getErr error
	getOut := output.CaptureOutput(func() {
		getErr = GetUser(getCmd, []string{employeeID}, true, "json")
	})

	require.NoError(t, getErr)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	gotIDFloat, ok := got[0]["employee_id"].(float64)
	require.True(t, ok, "employee_id should be a number in response")
	assert.Equal(t, employeeID, strconv.Itoa(int(gotIDFloat)))
	assert.Contains(t, got[0], "email")
}
