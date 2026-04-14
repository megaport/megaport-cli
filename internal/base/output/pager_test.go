//go:build !wasm
// +build !wasm

package output

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	t.Setenv("MEGAPORT_PAGER", "cat")

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

	// Use "cat" as the pager so output passes through unchanged.
	t.Setenv("MEGAPORT_PAGER", "cat")

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
	SetNoPager(false)
	assert.False(t, getNoPager())
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
		_ = RunWithPager(func() error {
			fmt.Print(lines)
			return nil
		})
	})
	// Output must still appear (fallback path).
	assert.Equal(t, lines, out)
}
