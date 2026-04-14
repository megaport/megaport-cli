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
	"sync/atomic"

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
//
// The pager value is fully controlled by the process environment, which is
// trusted to the same degree as the user running the process — consistent with
// how git(1) and gh(1) handle $GIT_PAGER / $PAGER. In automated or shared
// environments, pass --no-pager or set MEGAPORT_PAGER="" to disable paging.
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
// detection. Protected by atomic load/store so parallel tests do not race.
var terminalHeightOverride atomic.Int64

// setTerminalHeightForTesting overrides terminal height detection for tests.
// Pass 0 to restore automatic detection. Unexported to keep test scaffolding
// out of the public API while still being accessible within the package.
func setTerminalHeightForTesting(h int) {
	terminalHeightOverride.Store(int64(h))
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
	// Close before Remove so the file handle is released first (required on Windows).
	defer func() { _ = tmp.Close(); _ = os.Remove(tmp.Name()) }()

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
	height := int(terminalHeightOverride.Load())
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
	//nolint:gosec // pager command is user-controlled via MEGAPORT_PAGER/PAGER env vars,
	// which are trusted to the same degree as the calling user (see resolvePager).
	cmd := exec.Command(parts[0], parts[1:]...)
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
