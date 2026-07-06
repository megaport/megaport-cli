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
	// no-op passthrough in session_expired_native.go, even if
	// MEGAPORT_ACCESS_TOKEN is set (e.g. left over from a shared shell that
	// also ran WASM tests). A native process must never reclassify an ordinary
	// credential-auth failure into the WASM-only session-expired signal.
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	for _, statusCode := range []int{401, 403} {
		err := wrapSessionExpiredError(makeAPIError(statusCode, ""))
		assert.Equal(t, exitcodes.Authentication, classifyError(err))
		assert.NotContains(t, err.Error(), SessionExpiredMarker)
	}

	assert.Nil(t, wrapSessionExpiredError(nil))

	plain := errors.New("boom")
	assert.Equal(t, plain, wrapSessionExpiredError(plain))
}
