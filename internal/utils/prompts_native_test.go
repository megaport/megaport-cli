//go:build !js || !wasm
// +build !js !wasm

package utils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	defer inR.Close()
	defer inW.Close()

	outR, outW, err := os.Pipe()
	require.NoError(t, err)
	defer outR.Close()

	os.Stdin = inR
	os.Stdout = outW
	_ = inW.Close()

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, outR)
		close(done)
	}()

	for _, noColor := range []bool{true, false} {
		pw, err := nativePasswordPrompt("Enter password:", noColor)
		assert.Error(t, err, "expected error reading password from non-tty pipe")
		assert.Empty(t, pw)
	}

	_ = outW.Close()
	<-done
}
