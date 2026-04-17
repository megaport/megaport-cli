//go:build !js || !wasm
// +build !js !wasm

package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNativeInitAssignsPasswordPromptFn(t *testing.T) {
	assert.NotNil(t, GetPasswordPrompt(), "native init should wire passwordPromptFn")
}

func TestNativePasswordPrompt_NonTTYReturnsError(t *testing.T) {
	// term.ReadPassword fails when stdin is not a terminal. This test
	// explicitly replaces os.Stdin with a pipe to exercise that error path.
	oldStdin, oldStdout := os.Stdin, os.Stdout
	defer func() { os.Stdin, os.Stdout = oldStdin, oldStdout }()

	inR, inW, err := os.Pipe()
	assert.NoError(t, err)
	outR, outW, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdin = inR
	os.Stdout = outW
	_ = inW.Close()

	for _, noColor := range []bool{true, false} {
		pw, err := nativePasswordPrompt("Enter password:", noColor)
		assert.Error(t, err, "expected error reading password from non-tty pipe")
		assert.Empty(t, pw)
	}

	_ = outW.Close()
	_, _ = outR.Read(make([]byte, 1024))
	_ = inR.Close()
	_ = outR.Close()
}
