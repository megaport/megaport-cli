//go:build !wasm

package output

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTempFile only exists in the native build, so this test lives behind the
// !wasm tag alongside it.
func TestCaptureOutput_TempFileFailure(t *testing.T) {
	orig := createTempFile
	createTempFile = func() (*os.File, error) {
		return nil, errors.New("temp file unavailable")
	}
	defer func() { createTempFile = orig }()

	called := false
	result := CaptureOutput(func() { called = true })

	assert.True(t, called, "f should still be called when temp file creation fails")
	assert.Empty(t, result, "result should be empty when temp file creation fails")
}
