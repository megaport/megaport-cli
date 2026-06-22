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

// mveStatusFromSDK reads the MVE through the SDK client and returns its
// provisioning status. Like mveLockedFromSDK it uses its own short-lived
// context so a slow read can't exhaust a context shared with earlier calls.
func mveStatusFromSDK(t *testing.T, client *megaport.Client, uid string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	mve, err := client.MVEService.GetMVE(ctx, uid)
	require.NoErrorf(t, err, "SDK GetMVE failed for %s", uid)
	require.NotNil(t, mve)
	return mve.ProvisioningStatus
}

// buyArubaMVEForLifecycle provisions an Aruba MVE on staging via the buy CLI
// action (which waits for provisioning) and registers a best-effort hard delete
// in t.Cleanup. It returns the new MVE's UID. The image is discovered
// dynamically because staging image IDs change.
func buyArubaMVEForLifecycle(t *testing.T, namePrefix string) string {
	t.Helper()
	img := discoverArubaImage(t)
	name := fmt.Sprintf("%s-%s", namePrefix, generateUniqueID())

	vendorConfig := fmt.Sprintf(`{
		"vendor": "aruba",
		"productSize": "%s",
		"imageId": %d,
		"accountName": "test",
		"accountKey": "test",
		"systemTag": "test"
	}`, img.productSize(), img.ID)
	vnics := `[{"description": "MVE VNIC 1", "vlan": 55}, {"description": "MVE VNIC 2", "vlan": 56}]`

	buyCmd := integrationMVEBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(stagingMVELocationID)))
	require.NoError(t, buyCmd.Flags().Set("vendor-config", vendorConfig))
	require.NoError(t, buyCmd.Flags().Set("vnics", vnics))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMVE(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MVE output: %s", buyOut)

	mveUID := parseCreatedUID(buyOut, "MVE")
	// Register the delete cleanup before asserting on the UID, so a created MVE
	// is always torn down (CANCEL_NOW) even if the parse below fails.
	t.Cleanup(func() {
		if mveUID == "" {
			return
		}
		delCmd := integrationMVEDeleteCmd()
		_ = delCmd.Flags().Set("force", "true")
		var delErr error
		out := captureTableOutput(func() { delErr = DeleteMVE(delCmd, []string{mveUID}, true) })
		if delErr != nil {
			t.Logf("cleanup: delete MVE %s failed: %v; output: %s", mveUID, delErr, out)
			return
		}
		t.Logf("cleanup: delete MVE %s: %s", mveUID, out)
	})
	require.NotEmpty(t, mveUID, "could not parse MVE UID from: %s", buyOut)
	return mveUID
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

	mveUID := buyArubaMVEForLifecycle(t, "CLI-Test-MVE-Lock")

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

// TestIntegration_MVERestoreLifecycle buys an MVE, schedules it for cancellation
// (terminate-later), then restores it and asserts it is active again. Restore
// (UN_CANCEL) only works on a CANCELLED resource, and the CLI/SDK DeleteMVE
// forces CANCEL_NOW (immediate, non-restorable decommission), so the cancellation
// is scheduled through the SDK's DeleteProduct with DeleteNow=false. If staging
// doesn't leave the MVE in a CANCELLED state (no terminate-later window), the
// test skips with the observed status rather than flaking.
func TestIntegration_MVERestoreLifecycle(t *testing.T) {
	testutil.RequireStagingForProvisioning(t)
	client := testutil.SetupIntegrationClient(t)
	t.Cleanup(testutil.LoginWithClient(t, client))

	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	mveUID := buyArubaMVEForLifecycle(t, "CLI-Test-MVE-Restore")

	// Schedule cancellation (terminate-later) directly through the SDK, in its
	// own context. A failure here means staging won't schedule a cancellation for
	// this resource, so there's no restore window; the MVE is still live and the
	// delete cleanup removes it.
	cancelCtx, cancelCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancelCancel()
	if _, err := client.ProductService.DeleteProduct(cancelCtx, &megaport.DeleteProductRequest{
		ProductID: mveUID,
		DeleteNow: false,
	}); err != nil {
		t.Skipf("staging rejected terminate-later (CANCEL) for MVE %s, so there is no restore window: %v", mveUID, err)
	}

	// Confirm the MVE reached a restorable CANCELLED state. If staging
	// decommissioned it immediately (or hasn't reflected the cancel yet), there's
	// no window to restore from.
	if status := mveStatusFromSDK(t, client, mveUID); status != megaport.STATUS_CANCELLED {
		t.Skipf("restore needs a CANCELLED (terminate-later) MVE; staging left it %q after CANCEL", status)
	}

	// Restore (UN_CANCEL) through the CLI and assert the MVE is active again.
	var restoreErr error
	restoreOut := captureTableOutput(func() { restoreErr = RestoreMVE(&cobra.Command{Use: "restore"}, []string{mveUID}, true) })
	require.NoError(t, restoreErr, "restore MVE output: %s", restoreOut)

	status := mveStatusFromSDK(t, client, mveUID)
	assert.Containsf(t, megaport.SERVICE_STATE_READY, status,
		"MVE should be active (LIVE/CONFIGURED) after restore, got %q", status)
}
