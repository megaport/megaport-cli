//go:build !js && !wasm

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

// TestNativePasswordPrompt_UsesSharedReaderWhenDataBuffered is a regression
// test for the stdinHasBuffered guard: if an earlier prompt in the same flow
// already read ahead past its own line, nativePasswordPrompt must consume the
// buffered leftover through the shared reader rather than calling
// term.ReadPassword, which reads straight off the fd and would never see it.
func TestNativePasswordPrompt_UsesSharedReaderWhenDataBuffered(t *testing.T) {
	oldStdin, oldStdout := os.Stdin, os.Stdout
	defer func() {
		os.Stdin, os.Stdout = oldStdin, oldStdout
		resetSharedStdinReader()
	}()

	inR, inW, err := os.Pipe()
	require.NoError(t, err)
	defer inR.Close()

	outR, outW, err := os.Pipe()
	require.NoError(t, err)
	defer outR.Close()

	os.Stdin = inR
	os.Stdout = outW
	resetSharedStdinReader()

	_, err = inW.WriteString("first\nsecond\n")
	require.NoError(t, err)
	require.NoError(t, inW.Close())

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, outR)
		close(done)
	}()

	first, err := readStdinLine()
	require.NoError(t, err)
	assert.Equal(t, "first", first)

	// Read is allowed to return fewer bytes than requested, so the read above
	// isn't guaranteed to have pulled "second\n" into the buffer on its own.
	// Peek forces fill() to pull at least one more byte so the assertion
	// below is deterministic.
	stdinReaderMu.Lock()
	_, peekErr := stdinReader.Peek(1)
	stdinReaderMu.Unlock()
	require.NoError(t, peekErr)

	require.True(t, stdinHasBuffered())

	pw, err := nativePasswordPrompt("Enter password:", true)
	require.NoError(t, err)
	assert.Equal(t, "second", pw)

	_ = outW.Close()
	<-done
}
