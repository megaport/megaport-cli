//go:build integration && provisioning

package mcr

import (
	"context"
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

// mcrLockedFromSDK reads the MCR through the SDK client and returns its customer
// lock flag. The CLI's JSON output doesn't surface the locked field, so lock
// assertions go through the SDK directly.
func mcrLockedFromSDK(t *testing.T, client *megaport.Client, uid string) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	mcr, err := client.MCRService.GetMCR(ctx, uid)
	require.NoErrorf(t, err, "SDK GetMCR failed for %s", uid)
	require.NotNil(t, mcr)
	return mcr.Locked
}

// TestIntegration_MCRLockLifecycle buys an MCR, locks it, asserts the lock via an
// SDK read, unlocks it, and asserts the flag cleared. Cleanup unlocks before
// deleting so a failure mid-lock never leaves the resource stuck. It carries the
// extra `provisioning` build tag so the nightly read-only job never runs it; it
// runs in the manual provisioning job.
func TestIntegration_MCRLockLifecycle(t *testing.T) {
	testutil.RequireStagingForProvisioning(t)
	client := testutil.SetupIntegrationClient(t)
	// Restore login via t.Cleanup, not defer: defers run before t.Cleanup, so a
	// deferred restore would swap back the default login (wrong environment)
	// before the resource cleanup below gets to run.
	t.Cleanup(testutil.LoginWithClient(t, client))

	// Action functions mutate the process-wide output format; restore it so
	// test order can't leak state between tests in this package.
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	name := fmt.Sprintf("CLI-Test-MCR-Lock-%s", generateUniqueID())
	locationID := testutil.FindMCRTestLocation(t, client, 1000, stagingMCRLocationID)

	// BuyMCR waits for provisioning (no --no-wait), so the MCR is lockable once
	// it returns.
	buyCmd := integrationMCRBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(locationID)))
	require.NoError(t, buyCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut)
	// Register cleanup before asserting on mcrUID, so any created MCR is cleaned
	// up even if the UID parse fails. Unlock first: the API can refuse to delete
	// a locked resource, so a test that fails after locking must unlock here.
	t.Cleanup(func() {
		if mcrUID == "" {
			return
		}
		unlockOut := captureTableOutput(func() { _ = UnlockMCR(&cobra.Command{Use: "unlock"}, []string{mcrUID}, true) })
		t.Logf("cleanup: unlock MCR %s: %s", mcrUID, unlockOut)
		delCmd := integrationMCRDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMCR(delCmd, []string{mcrUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MCR %s failed: %v; output: %s", mcrUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MCR %s: %s", mcrUID, out)
	})
	require.NotEmpty(t, mcrUID, "could not parse MCR UID from: %s", buyOut)

	// A freshly provisioned MCR starts unlocked.
	require.False(t, mcrLockedFromSDK(t, client, mcrUID), "MCR should start unlocked")

	// Lock and assert the flag set via SDK read.
	var lockErr error
	lockOut := captureTableOutput(func() { lockErr = LockMCR(&cobra.Command{Use: "lock"}, []string{mcrUID}, true) })
	require.NoError(t, lockErr, "lock MCR output: %s", lockOut)
	assert.True(t, mcrLockedFromSDK(t, client, mcrUID), "MCR should be locked after lock")

	// Unlock and assert the flag cleared.
	var unlockErr error
	unlockOut := captureTableOutput(func() { unlockErr = UnlockMCR(&cobra.Command{Use: "unlock"}, []string{mcrUID}, true) })
	require.NoError(t, unlockErr, "unlock MCR output: %s", unlockOut)
	assert.False(t, mcrLockedFromSDK(t, client, mcrUID), "MCR should be unlocked after unlock")
}
