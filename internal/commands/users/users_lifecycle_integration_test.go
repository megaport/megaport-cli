//go:build integration && provisioning

package users

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 4)
	_, err := crypto_rand.Read(buf)
	require.NoError(t, err, "failed to read crypto/rand entropy")
	return hex.EncodeToString(buf)
}

func newCreateUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "create"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("position", "", "")
	cmd.Flags().String("phone", "", "")
	return cmd
}

func newUpdateUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("position", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Bool("notification-enabled", false, "")
	return cmd
}

func newDeleteUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().Bool("force", false, "")
	return cmd
}

// employeeIDByEmail polls the company user list for the user with the given
// email and returns its employee ID (PartyId, falling back to PersonId — the
// same derivation the get/list output uses). It retries briefly because a
// freshly created user can take a moment to appear in the list.
func employeeIDByEmail(t *testing.T, client *megaport.Client, email string) int {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		users, err := client.UserManagementService.ListCompanyUsers(ctx)
		cancel()
		require.NoError(t, err, "SDK ListCompanyUsers failed")
		for _, u := range users {
			if u == nil || u.Email != email {
				continue
			}
			id := u.PartyId
			if id == 0 {
				id = u.PersonId
			}
			require.NotZerof(t, id, "user %s has no usable employee ID", email)
			return id
		}
		if time.Now().After(deadline) {
			t.Fatalf("created user %s did not appear in company user list within timeout", email)
		}
		time.Sleep(2 * time.Second)
	}
}

func userExistsByEmail(t *testing.T, client *megaport.Client, email string) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	users, err := client.UserManagementService.ListCompanyUsers(ctx)
	require.NoError(t, err, "SDK ListCompanyUsers failed")
	for _, u := range users {
		if u != nil && u.Email == email {
			return true
		}
	}
	return false
}

// TestIntegration_UserLifecycle exercises the full create/get/update/delete path
// of the user CLI actions against staging. The invited user is created with an
// @example.com address (a reserved domain that never delivers mail) and is in a
// pending-invitation state, which is the only state in which a user can be
// deleted. A t.Cleanup safety net removes the user if the test fails before its
// own delete step. This test carries the extra `provisioning` build tag so the
// nightly read-only job never runs it; it runs in the manual provisioning job.
func TestIntegration_UserLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	suffix := uniqueSuffix(t)
	email := fmt.Sprintf("cli-test-%s@example.com", suffix)
	firstName := "CLITest"
	lastName := "User-" + suffix

	createCmd := newCreateUserCmd()
	require.NoError(t, createCmd.Flags().Set("first-name", firstName))
	require.NoError(t, createCmd.Flags().Set("last-name", lastName))
	require.NoError(t, createCmd.Flags().Set("email", email))
	require.NoError(t, createCmd.Flags().Set("position", "Read Only"))

	require.NoError(t, CreateUser(createCmd, nil, true), "CreateUser failed")

	employeeID := employeeIDByEmail(t, client, email)
	t.Logf("created user %s with employee ID %d", email, employeeID)

	// Safety net: remove the user even if a later step fails. Best-effort —
	// the happy path deletes the user itself, so this may find it already gone.
	t.Cleanup(func() {
		if !userExistsByEmail(t, client, email) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.UserManagementService.DeleteUser(ctx, employeeID); err != nil {
			t.Logf("cleanup: failed to delete user %d (%s): %v", employeeID, email, err)
		}
	})

	idArg := strconv.Itoa(employeeID)

	getCmd := &cobra.Command{Use: "get"}
	getOut := output.CaptureOutput(func() {
		require.NoError(t, GetUser(getCmd, []string{idArg}, true, "json"))
	})

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	assert.Equal(t, email, got[0]["email"])
	assert.Equal(t, firstName, got[0]["first_name"])

	newFirstName := "CLIUpdated"
	updateCmd := newUpdateUserCmd()
	require.NoError(t, updateCmd.Flags().Set("first-name", newFirstName))
	require.NoError(t, UpdateUser(updateCmd, []string{idArg}, true), "UpdateUser failed")

	updatedCtx, updatedCancel := context.WithTimeout(context.Background(), 30*time.Second)
	updated, err := client.UserManagementService.GetUser(updatedCtx, employeeID)
	updatedCancel()
	require.NoError(t, err, "SDK GetUser failed after update")
	require.NotNil(t, updated)
	assert.Equal(t, newFirstName, updated.FirstName, "first name should be updated")

	deleteCmd := newDeleteUserCmd()
	require.NoError(t, deleteCmd.Flags().Set("force", "true"))
	require.NoError(t, DeleteUser(deleteCmd, []string{idArg}, true), "DeleteUser failed")

	assert.False(t, userExistsByEmail(t, client, email), "user should be gone after delete")
}
