//go:build !js && !wasm
// +build !js,!wasm

package wasm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStubs exercises the no-op stub implementations used in non-WASM builds
// to ensure they compile and don't panic.
func TestStubs(t *testing.T) {
	ResetOutputBuffers()
	RegisterOutputStateReset(func() {})
	assert.Equal(t, "", GetCapturedOutput())
	SetupIO()

	called := false
	CaptureOutput(func() { called = true })
	assert.True(t, called)

	assert.Nil(t, SplitArgs("foo bar"))
}
