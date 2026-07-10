//go:build !js || !wasm

package utils

import (
	"errors"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/stretchr/testify/assert"
)

func TestWrapSessionExpiredError_NativeBuildNeverWraps(t *testing.T) {
	// wrapSessionExpiredError's WASM implementation only compiles into the WASM
	// binary (session_expired_wasm.go); this native binary always uses the
	// no-op passthrough in session_expired_native.go, even with the token-present
	// snapshot forced true. A native process must never reclassify an ordinary
	// credential-auth failure into the WASM-only session-expired signal.
	for _, statusCode := range []int{401, 403} {
		err := wrapSessionExpiredError(makeAPIError(statusCode, ""), true)
		assert.Equal(t, exitcodes.Authentication, classifyError(err))
		assert.NotContains(t, err.Error(), SessionExpiredMarker)
	}

	assert.Nil(t, wrapSessionExpiredError(nil, true))

	plain := errors.New("boom")
	assert.Equal(t, plain, wrapSessionExpiredError(plain, true))
}

func TestSessionTokenPresent_NativeAlwaysFalse(t *testing.T) {
	// The external-token path is WASM-only, so the snapshot is false on native
	// even when MEGAPORT_ACCESS_TOKEN happens to be set in the environment.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")
	assert.False(t, sessionTokenPresent())
}
