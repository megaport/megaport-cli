//go:build !wasm
// +build !wasm

package output

import (
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
// environments, pass --no-pager to disable paging.
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
//
// Terminal dimensions are read before os.Stdout is redirected so that table
// rendering (which calls getTerminalWidth via printTable) uses the real TTY
// width for column layout even while output is being buffered to a temp file.
// The temp file is streamed rather than loaded into memory, so paging large
// tables does not require holding the full output in RAM.
func RunWithPager(fn func() error) error {
	if !IsTerminal() || getNoPager() {
		return fn()
	}

	pagerMu.Lock()
	defer pagerMu.Unlock()

	origStdout := os.Stdout

	// Capture real terminal dimensions before swapping os.Stdout so that both
	// column-width layout and the height gate use the actual TTY geometry.
	height := int(terminalHeightOverride.Load())
	if height <= 0 {
		w, h, sizeErr := term.GetSize(int(origStdout.Fd()))
		if sizeErr == nil {
			if h > 0 {
				height = h
			}
			if w > 0 {
				// Override width so printTable uses the real TTY width while
				// os.Stdout is redirected to the temp file.
				terminalWidthOverride.Store(int64(w))
				defer terminalWidthOverride.Store(0)
			}
		}
	}

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
	defer func() {
		os.Stdout = origStdout
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	fnErr := fn()

	// Count lines by scanning the file incrementally — avoids loading the
	// entire output into memory and has no token-size limit.
	lineCount, err := countLines(tmp)
	if err != nil {
		// If counting fails, fall back to direct write so output is not lost.
		if _, seekErr := tmp.Seek(0, io.SeekStart); seekErr == nil {
			_, _ = io.Copy(origStdout, tmp)
		}
		return fnErr
	}
	if lineCount == 0 {
		// fn() may have written content without a trailing newline (countLines
		// counts '\n' only). Check the file size: if non-zero, treat the
		// partial last line as a single line rather than dropping it.
		info, statErr := tmp.Stat()
		if statErr != nil {
			// Cannot determine size; fall back to direct write.
			// countLines already rewound on success but failed here, so seek manually.
			if _, seekErr := tmp.Seek(0, io.SeekStart); seekErr == nil {
				_, _ = io.Copy(origStdout, tmp)
			}
			return fnErr
		}
		if info.Size() == 0 {
			return fnErr
		}
		lineCount = 1
	}

	// countLines rewinds r to offset 0 on success, so tmp is ready to copy.
	if height <= 0 {
		// Terminal height unknown; write directly.
		_, _ = io.Copy(origStdout, tmp)
		return fnErr
	}

	if lineCount <= height {
		_, _ = io.Copy(origStdout, tmp)
		return fnErr
	}

	// Output exceeds terminal height: pipe through pager.
	if err := runPager(resolvePager(), tmp, origStdout); err != nil {
		// Pager failed; seek back and fall back to direct write so output is not lost.
		if _, seekErr := tmp.Seek(0, io.SeekStart); seekErr == nil {
			_, _ = io.Copy(origStdout, tmp)
		}
	}
	return fnErr
}

// countLines counts the number of newline characters in r by reading in
// chunks. It seeks r to the start before counting and back to the start
// before returning, so callers always receive r positioned at offset 0.
// Unlike bufio.Scanner this approach has no token-size limit and handles
// arbitrarily wide table rows.
func countLines(r io.ReadSeeker) (int, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	count := 0
	buf := make([]byte, 32*1024)
	for {
		n, readErr := r.Read(buf)
		for _, b := range buf[:n] {
			if b == '\n' {
				count++
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return count, readErr
		}
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return count, err
	}
	return count, nil
}

// runPager spawns pagerCmd, streams content to its stdin, and waits for it to exit.
func runPager(pagerCmd string, content io.Reader, stdout *os.File) error {
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
	_, _ = io.Copy(stdin, content)
	_ = stdin.Close()
	return cmd.Wait()
}
