//go:build !wasm
// +build !wasm

package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/term"
)

var (
	noPagerMu  sync.RWMutex
	noPagerVal bool
)

// SetNoPager disables or enables the pager. When true, output is always written
// directly to stdout even if it exceeds the terminal height.
func SetNoPager(v bool) {
	noPagerMu.Lock()
	defer noPagerMu.Unlock()
	noPagerVal = v
}

func getNoPager() bool {
	noPagerMu.RLock()
	defer noPagerMu.RUnlock()
	return noPagerVal
}

// resolvePager returns the pager command to use.
// Precedence: MEGAPORT_PAGER > PAGER > "less -R".
func resolvePager() string {
	if v := os.Getenv("MEGAPORT_PAGER"); v != "" {
		return v
	}
	if v := os.Getenv("PAGER"); v != "" {
		return v
	}
	return "less -R"
}

// terminalHeightOverride, when > 0, replaces automatic terminal height
// detection. Set via SetTerminalHeightForTesting; reset by passing 0.
var terminalHeightOverride int

// SetTerminalHeightForTesting overrides terminal height detection for tests.
// Pass 0 to restore automatic detection.
func SetTerminalHeightForTesting(h int) {
	terminalHeightOverride = h
}

// pagerMu serialises RunWithPager calls. A separate mutex from stdoutMu is
// used intentionally so that CaptureOutput (which holds stdoutMu) can safely
// call code that goes through RunWithPager without deadlocking.
var pagerMu sync.Mutex

// RunWithPager runs fn, capturing its stdout output. If stdout is a TTY, the
// pager is not disabled, and the captured line count exceeds the terminal
// height, the output is piped through the configured pager. Otherwise the
// output is written directly to stdout.
func RunWithPager(fn func() error) error {
	if !IsTerminal() || getNoPager() {
		return fn()
	}

	pagerMu.Lock()
	defer pagerMu.Unlock()

	origStdout := os.Stdout

	// Use a temp file to buffer output. An os.Pipe would deadlock once the
	// pipe buffer fills up for large tables, while a temp file does not.
	tmp, err := os.CreateTemp("", "pager-*")
	if err != nil {
		// Cannot create temp file; render directly.
		return fn()
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	os.Stdout = tmp
	fnErr := fn()
	os.Stdout = origStdout

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fnErr
	}
	content, err := io.ReadAll(tmp)
	if err != nil || len(content) == 0 {
		return fnErr
	}

	// Determine terminal height.
	height := terminalHeightOverride
	if height <= 0 {
		_, h, sizeErr := term.GetSize(int(origStdout.Fd()))
		if sizeErr != nil || h <= 0 {
			// Cannot determine terminal size; write directly.
			_, _ = origStdout.Write(content)
			return fnErr
		}
		height = h
	}

	if bytes.Count(content, []byte("\n")) <= height {
		_, _ = origStdout.Write(content)
		return fnErr
	}

	// Output exceeds terminal height: pipe through pager.
	if err := runPager(resolvePager(), content, origStdout); err != nil {
		// Pager failed; fall back to direct write so output is not lost.
		_, _ = origStdout.Write(content)
	}
	return fnErr
}

// runPager spawns pagerCmd, pipes content to its stdin, and waits for it to exit.
func runPager(pagerCmd string, content []byte, stdout *os.File) error {
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty pager command")
	}
	cmd := exec.Command(parts[0], parts[1:]...) //nolint:gosec // pager command comes from trusted env var / default
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	_, _ = stdin.Write(content)
	_ = stdin.Close()
	return cmd.Wait()
}
