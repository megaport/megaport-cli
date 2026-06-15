//go:build integration && provisioning

package users

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

func newDeleteUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().Bool("force", false, "")
	return cmd
}

// findUserIDByEmail looks up the user with the given email and returns its
// employee ID (PartyId, falling back to PersonId — the same derivation the
// get/list output uses), with ok=false if no such user is present. It returns
// the list error rather than aborting, so it is safe to call from a t.Cleanup
// callback, where a FailNow would mask the original failure.
func findUserIDByEmail(client *megaport.Client, email string) (id int, ok bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	users, err := client.UserManagementService.ListCompanyUsers(ctx)
	if err != nil {
		return 0, false, err
	}
	for _, u := range users {
		// Email match is case-insensitive: the API may normalize the
		// local-part casing of the address we sent.
		if u == nil || !strings.EqualFold(u.Email, email) {
			continue
		}
		id := u.PartyId
		if id == 0 {
			id = u.PersonId
		}
		return id, true, nil
	}
	return 0, false, nil
}

// employeeIDByEmail polls until the user appears in the company user list and
// returns its employee ID. It retries briefly because a freshly created user
// can take a moment to appear.
func employeeIDByEmail(t *testing.T, client *megaport.Client, email string) int {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		id, ok, err := findUserIDByEmail(client, email)
		require.NoError(t, err, "SDK ListCompanyUsers failed")
		if ok {
			require.NotZerof(t, id, "user %s has no usable employee ID", email)
			return id
		}
		if time.Now().After(deadline) {
			t.Fatalf("created user %s did not appear in company user list within timeout", email)
		}
		time.Sleep(2 * time.Second)
	}
}

// TestIntegration_UserLifecycle exercises the create/get/delete path of the user
// CLI actions against staging. The invited user is created with an
// @sink.megaport.com address (the sink domain the staging account requires; mail
// to it is never delivered) and stays in a pending-invitation state. Update is
// deliberately not exercised: staging rejects updates to a pending user, and a
// test can't accept an emailed invitation. Delete, by contrast, is only allowed
// while the invitation is pending. A t.Cleanup safety net removes the user if the
// test fails before its own delete step. This test carries the extra
// `provisioning` build tag so the nightly read-only job never runs it; it runs in
// the manual provisioning job.
func TestIntegration_UserLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	suffix := uniqueSuffix(t)
	// CLI-Test- prefix matches the convention other provisioning tests use so
	// leftover staging artifacts are easy to find.
	email := fmt.Sprintf("CLI-Test-%s@sink.megaport.com", suffix)
	firstName := "CLITest"
	lastName := "User-" + suffix

	createCmd := newCreateUserCmd()
	require.NoError(t, createCmd.Flags().Set("first-name", firstName))
	require.NoError(t, createCmd.Flags().Set("last-name", lastName))
	require.NoError(t, createCmd.Flags().Set("email", email))
	require.NoError(t, createCmd.Flags().Set("position", "Read Only"))

	require.NoError(t, CreateUser(createCmd, nil, true), "CreateUser failed")

	// Safety net registered before the ID lookup below can fail: if CreateUser
	// succeeded but employeeIDByEmail times out, this still removes the user.
	// It resolves the ID by email at cleanup time and is best-effort — the happy
	// path deletes the user itself, so this often finds it already gone.
	t.Cleanup(func() {
		id, ok, err := findUserIDByEmail(client, email)
		if err != nil {
			t.Logf("cleanup: could not list users to find %s: %v", email, err)
			return
		}
		if !ok {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.UserManagementService.DeleteUser(ctx, id); err != nil {
			t.Logf("cleanup: failed to delete user %d (%s): %v", id, email, err)
		}
	})

	employeeID := employeeIDByEmail(t, client, email)
	t.Logf("created user %s with employee ID %d", email, employeeID)

	idArg := strconv.Itoa(employeeID)

	getCmd := &cobra.Command{Use: "get"}
	getOut := output.CaptureOutput(func() {
		require.NoError(t, GetUser(getCmd, []string{idArg}, true, "json"))
	})

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	gotEmail, _ := got[0]["email"].(string)
	assert.Truef(t, strings.EqualFold(email, gotEmail), "email mismatch: want %q got %q", email, gotEmail)
	assert.Equal(t, firstName, got[0]["first_name"])

	deleteCmd := newDeleteUserCmd()
	require.NoError(t, deleteCmd.Flags().Set("force", "true"))
	require.NoError(t, DeleteUser(deleteCmd, []string{idArg}, true), "DeleteUser failed")

	_, stillExists, err := findUserIDByEmail(client, email)
	require.NoError(t, err, "SDK ListCompanyUsers failed")
	assert.False(t, stillExists, "user should be gone after delete")
}
