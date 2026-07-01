//go:build integration && provisioning

package mve

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

// mveLockedFromSDK reads the MVE through the SDK client and returns its customer
// lock flag. The CLI's JSON output doesn't surface the locked field, so lock
// assertions go through the SDK directly.
func mveLockedFromSDK(t *testing.T, client *megaport.Client, uid string) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	mve, err := client.MVEService.GetMVE(ctx, uid)
	require.NoErrorf(t, err, "SDK GetMVE failed for %s", uid)
	require.NotNil(t, mve)
	return mve.Locked
}

// buyArubaMVEForLifecycle provisions an Aruba MVE on staging via the buy CLI
// action (which waits for provisioning) and registers a best-effort hard delete
// in t.Cleanup. It returns the new MVE's UID. The image is discovered
// dynamically because staging image IDs change.
func buyArubaMVEForLifecycle(t *testing.T, client *megaport.Client, namePrefix string) string {
	t.Helper()
	img := discoverArubaImage(t)
	name := fmt.Sprintf("%s-%s", namePrefix, generateUniqueID())
	vnics := `[{"description": "MVE VNIC 1", "vlan": 55}, {"description": "MVE VNIC 2", "vlan": 56}]`

	return buyMVEAtAvailableLocation(t, client, img, func(locationID int, size string) *cobra.Command {
		vendorConfig := fmt.Sprintf(`{
			"vendor": "aruba",
			"productSize": "%s",
			"imageId": %d,
			"accountName": "test",
			"accountKey": "test",
			"systemTag": "test"
		}`, size, img.ID)
		buyCmd := integrationMVEBuyCmd()
		require.NoError(t, buyCmd.Flags().Set("name", name))
		require.NoError(t, buyCmd.Flags().Set("term", "1"))
		require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(locationID)))
		require.NoError(t, buyCmd.Flags().Set("vendor-config", vendorConfig))
		require.NoError(t, buyCmd.Flags().Set("vnics", vnics))
		require.NoError(t, buyCmd.Flags().Set("yes", "true"))
		return buyCmd
	})
}

// TestIntegration_MVELockLifecycle buys an MVE, locks it, asserts the lock via an
// SDK read, unlocks it, and asserts the flag cleared. Cleanup unlocks before
// deleting so a failure mid-lock never leaves the resource stuck. It carries the
// extra `provisioning` build tag so the nightly read-only job never runs it; it
// runs in the manual provisioning job.
func TestIntegration_MVELockLifecycle(t *testing.T) {
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

	mveUID := buyArubaMVEForLifecycle(t, client, "CLI-Test-MVE-Lock")

	// Registered after the delete cleanup so it runs first (cleanups run LIFO):
	// the API can refuse to delete a locked resource, so unlock before delete.
	t.Cleanup(func() {
		unlockOut := captureTableOutput(func() { _ = UnlockMVE(&cobra.Command{Use: "unlock"}, []string{mveUID}, true) })
		t.Logf("cleanup: unlock MVE %s: %s", mveUID, unlockOut)
	})

	// A freshly provisioned MVE starts unlocked.
	require.False(t, mveLockedFromSDK(t, client, mveUID), "MVE should start unlocked")

	// Lock and assert the flag set via SDK read.
	var lockErr error
	lockOut := captureTableOutput(func() { lockErr = LockMVE(&cobra.Command{Use: "lock"}, []string{mveUID}, true) })
	require.NoError(t, lockErr, "lock MVE output: %s", lockOut)
	assert.True(t, mveLockedFromSDK(t, client, mveUID), "MVE should be locked after lock")

	// Unlock and assert the flag cleared.
	var unlockErr error
	unlockOut := captureTableOutput(func() { unlockErr = UnlockMVE(&cobra.Command{Use: "unlock"}, []string{mveUID}, true) })
	require.NoError(t, unlockErr, "unlock MVE output: %s", unlockOut)
	assert.False(t, mveLockedFromSDK(t, client, mveUID), "MVE should be unlocked after unlock")
}
