//go:build js && wasm

package utils

import (
	"errors"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapSessionExpiredError_TokenPathAuthFailure(t *testing.T) {
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	for _, statusCode := range []int{401, 403} {
		wrapped := wrapSessionExpiredError(makeAPIError(statusCode, ""))
		assert.Contains(t, wrapped.Error(), SessionExpiredMarker)
		assert.Equal(t, exitcodes.SessionExpired, classifyError(wrapped))
	}
}

func TestWrapSessionExpiredError_TokenPathNonAuthFailure(t *testing.T) {
	// A 500 on the token path is a normal API error, not a rejected token.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	err := wrapSessionExpiredError(makeAPIError(500, ""))
	assert.NotContains(t, err.Error(), SessionExpiredMarker)
	assert.Equal(t, exitcodes.API, classifyError(err))
}

func TestWrapSessionExpiredError_NilAndNonSDKErrors(t *testing.T) {
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	assert.Nil(t, wrapSessionExpiredError(nil))

	plain := errors.New("boom")
	assert.Equal(t, plain, wrapSessionExpiredError(plain))
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

func TestWrapRunE_SessionExpired_SurvivesActionsOwnPrintError(t *testing.T) {
	// Several actions call output.PrintError themselves before returning the
	// error (e.g. users_actions.go's GetUser), which latches ErrorEmitted and
	// makes finishWithError skip its own PrintError call. The marker must
	// still end up in the error finishWithError returns, since that's what
	// ExecuteWithArgs renders into captured output on the WASM error path,
	// independent of whether PrintError already ran.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

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

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.SessionExpired, cliErr.Code)
}
