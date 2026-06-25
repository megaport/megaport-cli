//go:build integration && provisioning

package mcr

import (
	"context"
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

// integrationMCRIPSecAddOnCmd builds a bare cobra command carrying the flags the
// add/update IPSec add-on actions read, following the same bare-command pattern
// as the integration command builders in mcr_integration_test.go.
func integrationMCRIPSecAddOnCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "ipsec-addon"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Int("tunnel-count", 0, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

// waitForIPSecAddOnUID polls the MCR until an IPSec add-on surfaces and returns
// its UID, or "" if none appears before the timeout. The add command does not
// wait for provisioning, so the add-on UID the update/disable steps need only
// becomes readable once the API has registered it.
func waitForIPSecAddOnUID(t *testing.T, client *megaport.Client, mcrUID string) string {
	t.Helper()
	const (
		timeout  = 5 * time.Minute
		interval = 15 * time.Second
	)
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return ""
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		mcr, err := client.MCRService.GetMCR(ctx, mcrUID)
		cancel()
		if err != nil {
			// A transient API error mid-poll shouldn't fail the test; retry
			// until the deadline.
			t.Logf("transient GetMCR error while waiting for IPSec add-on (will retry): %v", err)
			time.Sleep(interval)
			continue
		}

		for _, addOn := range mcr.AddOns {
			// MCR.AddOns is concretely []*MCRAddOnIPsecConfig, so a non-empty
			// AddOnUID is the IPSec add-on we just requested.
			if addOn != nil && addOn.AddOnUID != "" {
				return addOn.AddOnUID
			}
		}
		time.Sleep(interval)
	}
}

// isIPSecAddOnUnavailable reports whether err from adding the add-on indicates
// staging does not support the IPSec add-on on this MCR, as opposed to a real
// request-building or SDK regression (which must fail the test, not skip it).
func isIPSecAddOnUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// Keep these markers narrow to capability absence. Generic authorization
	// language ("not enabled", "not permitted") and transient infrastructure
	// errors ("not available", e.g. "service not available") are deliberately
	// excluded so a real request-building, transient, or SDK regression still
	// fails the test rather than being skipped.
	for _, marker := range []string{
		"not supported", "unsupported", "not eligible",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// waitForMCRReadyBestEffort lets the MCR settle back into a ready state between
// add-on mutations. Mutating an add-on can briefly move the MCR out of ready,
// and a follow-up mutation can be rejected while it settles. Best-effort: a
// genuine problem surfaces on the next operation, so a timeout only warns. The
// context is padded one minute past WaitForMCRReady's own timeout so that
// timeout wins with a clear error rather than a context-cancelled one.
func waitForMCRReadyBestEffort(t *testing.T, client *megaport.Client, mcrUID, stage string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	if err := client.MCRService.WaitForMCRReady(ctx, mcrUID, 5*time.Minute); err != nil {
		t.Logf("warning: MCR %s did not return to ready before IPSec %s: %v", mcrUID, stage, err)
	}
}

// TestIntegration_MCRIPSecAddOnLifecycle provisions an MCR, adds an IPSec
// add-on, updates its tunnel count, then disables it, tearing the MCR down in
// t.Cleanup. It skips when staging does not support the add-on on the test MCR.
func TestIntegration_MCRIPSecAddOnLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	// Restore login via t.Cleanup, not defer: defers run before t.Cleanup, so a
	// deferred restore would swap back the default login (wrong environment)
	// before the MCR-deletion cleanup below gets to run.
	t.Cleanup(testutil.LoginWithClient(t, client))

	// Action functions mutate the process-wide output format; restore it so
	// test order can't leak state between tests in this package.
	origFmt := output.GetOutputFormat()
	t.Cleanup(func() { output.SetOutputFormat(origFmt) })

	name := fmt.Sprintf("CLI-Test-MCR-IPSec-%s", generateUniqueID())

	// Buy a new MCR using flags. BuyMCR waits for provisioning (no --no-wait),
	// so the MCR is ready for add-on operations once it returns.
	buyCmd := integrationMCRBuyCmd()
	require.NoError(t, buyCmd.Flags().Set("name", name))
	require.NoError(t, buyCmd.Flags().Set("term", "1"))
	require.NoError(t, buyCmd.Flags().Set("port-speed", "1000"))
	require.NoError(t, buyCmd.Flags().Set("location-id", strconv.Itoa(stagingMCRLocationID)))
	require.NoError(t, buyCmd.Flags().Set("marketplace-visibility", "false"))
	require.NoError(t, buyCmd.Flags().Set("yes", "true"))

	var buyErr error
	buyOut := captureTableOutput(func() { buyErr = BuyMCR(buyCmd, nil, true) })
	require.NoError(t, buyErr, "buy MCR output: %s", buyOut)

	mcrUID := parseCreatedUID(buyOut, "MCR")
	// Register cleanup before asserting on mcrUID, so any created MCR is
	// deleted even if the UID parse fails.
	t.Cleanup(func() {
		if mcrUID == "" {
			return
		}
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

	// Add an IPSec add-on with an explicit tunnel count. Skip only when staging
	// signals the add-on is unsupported on this MCR; any other error fails the
	// test, so a regression in the request building or SDK call is still caught.
	addCmd := integrationMCRIPSecAddOnCmd()
	// 10 and 20 are both valid IPSec tunnel counts (the SDK accepts 10, 20, 30).
	require.NoError(t, addCmd.Flags().Set("tunnel-count", "10"))
	var addErr error
	addOut := captureTableOutput(func() { addErr = AddMCRIPSecAddOn(addCmd, []string{mcrUID}, true) })
	if isIPSecAddOnUnavailable(addErr) {
		t.Skipf("IPSec add-on not supported on staging MCR %s: %v", mcrUID, addErr)
	}
	require.NoError(t, addErr, "add IPSec add-on output: %s", addOut)
	assert.Contains(t, addOut, "IPSec add-on added successfully to MCR:")

	// The add command does not wait for provisioning, so poll until the add-on
	// surfaces. It has been accepted by the time AddMCRIPSecAddOn returns without
	// error, so a missing add-on here is a real fault rather than an unsupported one.
	addOnUID := waitForIPSecAddOnUID(t, client, mcrUID)
	require.NotEmptyf(t, addOnUID, "IPSec add-on did not surface on MCR %s after a successful add", mcrUID)

	waitForMCRReadyBestEffort(t, client, mcrUID, "update")

	// Update the add-on to a different tunnel count.
	updCmd := integrationMCRIPSecAddOnCmd()
	require.NoError(t, updCmd.Flags().Set("tunnel-count", "20"))
	var updErr error
	updOut := captureTableOutput(func() {
		updErr = UpdateMCRIPSecAddOn(updCmd, []string{mcrUID, addOnUID}, true)
	})
	require.NoError(t, updErr, "update IPSec add-on output: %s", updOut)
	assert.Contains(t, updOut, "updated successfully - tunnel count: 20")

	waitForMCRReadyBestEffort(t, client, mcrUID, "disable")

	// Disable the add-on (tunnel-count 0).
	disableCmd := integrationMCRIPSecAddOnCmd()
	require.NoError(t, disableCmd.Flags().Set("tunnel-count", "0"))
	var disableErr error
	disableOut := captureTableOutput(func() {
		disableErr = UpdateMCRIPSecAddOn(disableCmd, []string{mcrUID, addOnUID}, true)
	})
	require.NoError(t, disableErr, "disable IPSec add-on output: %s", disableOut)
	assert.Contains(t, disableOut, "IPSec add-on disabled successfully")
}
