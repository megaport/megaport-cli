//go:build js && wasm

package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapSessionExpiredError_TokenPathAuthFailure(t *testing.T) {
	for _, statusCode := range []int{401, 403} {
		wrapped := wrapSessionExpiredError(makeAPIError(statusCode, ""), true)
		assert.Contains(t, wrapped.Error(), SessionExpiredMarker)
		assert.Equal(t, exitcodes.SessionExpired, classifyError(wrapped))
	}
}

func TestWrapSessionExpiredError_TokenPathNonAuthFailure(t *testing.T) {
	// A 500 on the token path is a normal API error, not a rejected token.
	err := wrapSessionExpiredError(makeAPIError(500, ""), true)
	assert.NotContains(t, err.Error(), SessionExpiredMarker)
	assert.Equal(t, exitcodes.API, classifyError(err))
}

func TestWrapSessionExpiredError_NotOnTokenPath(t *testing.T) {
	// With no external token in play (tokenPresent false), a 401 stays an
	// ordinary auth error: an OAuth/credential login must not be reclassified.
	err := wrapSessionExpiredError(makeAPIError(401, ""), false)
	assert.NotContains(t, err.Error(), SessionExpiredMarker)
	assert.Equal(t, exitcodes.Authentication, classifyError(err))
}

func TestWrapSessionExpiredError_NilAndNonSDKErrors(t *testing.T) {
	assert.Nil(t, wrapSessionExpiredError(nil, true))

	plain := errors.New("boom")
	assert.Equal(t, plain, wrapSessionExpiredError(plain, true))
}

func TestWrapRunE_SessionExpired_TokenPath(t *testing.T) {
	// End to end: a WASM token-path 401 surfaces the marker and exit code
	// through the same RunE wrapper native commands use.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		return makeAPIError(401, "")
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), SessionExpiredMarker)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.SessionExpired, cliErr.Code)
}

func TestWrapRunE_SessionExpired_UsesEntrySnapshotNotLiveEnv(t *testing.T) {
	// The token-path decision is snapshotted at wrapper entry, not re-read at
	// error-handling time. A concurrent command clearing MEGAPORT_ACCESS_TOKEN
	// while this request was in flight (simulated here by clearing it inside the
	// action) must not suppress the marker: the request really did fail on the
	// token this command started with.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		os.Unsetenv("MEGAPORT_ACCESS_TOKEN")
		return makeAPIError(401, "")
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), SessionExpiredMarker,
		"marker must reflect the token present at command start, not the live env at error time")
}

func TestWrapRunE_SessionExpired_SurvivesActionsOwnPrintError(t *testing.T) {
	// Several actions call output.PrintError themselves before returning the
	// error (e.g. users_actions.go's GetUser), which latches ErrorEmitted and
	// makes finishWithError skip its own PrintError call. wrapSessionExpiredError
	// still rewrites the error to carry the marker, so finishWithError must
	// re-emit it: the raw error the action printed has no marker, and under a
	// live-output handler that chunk has already streamed and can't be retracted.
	// Assert the marker reaches the captured output buffer (what the host reads),
	// not just the returned Go error value.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")
	wasm.ResetOutputBuffers()

	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		apiErr := makeAPIError(401, "")
		output.PrintError("Failed to get user: %v", false, apiErr)
		return apiErr
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)
	assert.True(t, output.ErrorEmitted())
	assert.Contains(t, err.Error(), SessionExpiredMarker)
	assert.Contains(t, wasm.WasmOutputBuffer.String(), SessionExpiredMarker,
		"marker must reach the captured output the host reads, not only the returned error")

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.SessionExpired, cliErr.Code)
}
