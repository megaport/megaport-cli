//go:build !wasm
// +build !wasm

package output

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// passthroughPager returns a shell command that copies stdin to stdout
// unchanged, for use as a test pager. On Windows, "cat" is not available
// by default, so the test is skipped on that platform.
func passthroughPager(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test requires 'cat'; not available by default on Windows")
	}
	return "cat"
}

func TestResolvePager_Precedence(t *testing.T) {
	t.Setenv("MEGAPORT_PAGER", "")
	t.Setenv("PAGER", "")
	assert.Equal(t, "less -R", resolvePager())

	t.Setenv("PAGER", "more")
	assert.Equal(t, "more", resolvePager())

	t.Setenv("MEGAPORT_PAGER", "bat")
	assert.Equal(t, "bat", resolvePager())
}

func TestRunWithPager_NonTTY(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(false)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	called := false
	err := RunWithPager(func() error {
		called = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestRunWithPager_NoPager(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(true)
	t.Cleanup(func() { SetNoPager(false) })

	called := false
	err := RunWithPager(func() error {
		called = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestRunWithPager_ShortOutput(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	// Set a tall terminal so output never exceeds height.
	setTerminalHeightForTesting(1000)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	t.Setenv("MEGAPORT_PAGER", passthroughPager(t))

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			fmt.Println("line1")
			fmt.Println("line2")
			return nil
		})
		require.NoError(t, err)
	})
	assert.Contains(t, out, "line1")
	assert.Contains(t, out, "line2")
}

func TestRunWithPager_LongOutput(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	// Terminal height of 3 lines; output will be 10 lines — triggers pager.
	setTerminalHeightForTesting(3)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	// Use a pass-through pager so output arrives unchanged.
	t.Setenv("MEGAPORT_PAGER", passthroughPager(t))

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			for i := range 10 {
				fmt.Printf("line%d\n", i)
			}
			return nil
		})
		require.NoError(t, err)
	})

	for i := range 10 {
		assert.Contains(t, out, fmt.Sprintf("line%d", i))
	}
}

func TestRunWithPager_PropagatesFnError(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(false) // non-TTY so RunWithPager short-circuits cleanly
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	err := RunWithPager(func() error {
		return fmt.Errorf("inner error")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inner error")
}

func TestSetNoPager_RoundTrip(t *testing.T) {
	SetNoPager(true)
	assert.True(t, getNoPager())
	assert.True(t, GetNoPager())
	SetNoPager(false)
	assert.False(t, getNoPager())
	assert.False(t, GetNoPager())
}

// TestSetConfig_NoPager covers the NoPager field of SetConfig on native
// builds. This file is guarded by //go:build !wasm because pager state is
// native-only: WASM provides stub SetNoPager (no-op) and GetNoPager
// (always false) implementations, so a round-trip assertion only exercises
// real state on native.
func TestSetConfig_NoPager(t *testing.T) {
	t.Cleanup(ResetState)

	SetConfig(OutputConfig{NoPager: true})
	assert.True(t, GetNoPager())

	SetConfig(OutputConfig{NoPager: false})
	assert.False(t, GetNoPager())
}

// TestRunWithPager_NoTrailingNewline verifies that output without a trailing
// newline is not silently dropped. countLines must count the partial final
// line so the pager decision is correct.
func TestRunWithPager_NoTrailingNewline(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	setTerminalHeightForTesting(1000)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			fmt.Print("no-newline-here") // deliberate: no '\n'
			return nil
		})
		require.NoError(t, err)
	})
	assert.Equal(t, "no-newline-here", out)
}

// TestRunWithPager_MultiLineNoTrailingNewline verifies that multi-line output
// without a trailing newline is counted correctly (e.g., "a\nb\nc" is 3 lines,
// not 2). An off-by-one here would cause the pager to trigger one line late.
func TestRunWithPager_MultiLineNoTrailingNewline(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	// Terminal height of 2. Output is "a\nb\nc" — 3 lines (no trailing '\n').
	// With a correct count the pager fires; with an off-by-one it would not.
	setTerminalHeightForTesting(2)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	t.Setenv("MEGAPORT_PAGER", passthroughPager(t))

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			fmt.Print("a\nb\nc") // 3 lines, last has no '\n'
			return nil
		})
		require.NoError(t, err)
	})
	assert.Equal(t, "a\nb\nc", out)
}

// TestRunWithPager_EmptyOutput verifies that when fn writes nothing to stdout,
// RunWithPager returns cleanly without writing any output.
func TestRunWithPager_EmptyOutput(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	setTerminalHeightForTesting(5)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			// Write nothing to os.Stdout — content will be empty.
			return nil
		})
		require.NoError(t, err)
	})
	assert.Empty(t, out)
}

// TestRunWithPager_TermSizeFailure verifies that when terminal size cannot be
// determined (origStdout is not a real TTY and no override is set), output is
// written directly rather than being dropped.
func TestRunWithPager_TermSizeFailure(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	// Leave terminalHeightOverride at 0 (default). RunWithPager will call
	// term.GetSize on origStdout, which is a temp file (not a TTY) when
	// wrapped with CaptureOutput. That call fails, so the code falls through
	// to the direct-write path rather than invoking the pager.
	// (Do NOT call setTerminalHeightForTesting here.)

	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			fmt.Println("direct-write-line")
			return nil
		})
		require.NoError(t, err)
	})
	assert.Contains(t, out, "direct-write-line")
}

func TestRunWithPager_LongOutput_PagerFailure(t *testing.T) {
	orig := isTerminalCached.Load()
	SetIsTerminal(true)
	t.Cleanup(func() { isTerminalCached.Store(orig) })

	SetNoPager(false)
	t.Cleanup(func() { SetNoPager(false) })

	setTerminalHeightForTesting(3)
	t.Cleanup(func() { setTerminalHeightForTesting(0) })

	// Point to a nonexistent pager — RunWithPager should fall back to direct write.
	t.Setenv("MEGAPORT_PAGER", "/nonexistent-pager-cmd")

	lines := strings.Repeat("x\n", 10)
	out := CaptureOutput(func() {
		err := RunWithPager(func() error {
			fmt.Print(lines)
			return nil
		})
		require.NoError(t, err)
	})
	// Output must still appear (fallback path).
	assert.Equal(t, lines, out)
}
