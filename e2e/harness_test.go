//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runTimeout bounds a single CLI invocation so a hung command fails the test
// rather than the whole suite.
const runTimeout = 30 * time.Second

// Result is the outcome of one CLI invocation.
type Result struct {
	Stdout string
	Stderr string
	Exit   int
}

// Run execs the CLI binary with args in a sandboxed environment (HOME pointed at
// a temp dir, a minimal PATH, nothing else) and returns its stdout, stderr, and
// exit code. Hermetic tests use this.
func Run(t *testing.T, args ...string) Result {
	t.Helper()
	return run(t, nil, args)
}

// RunWithEnv is like Run but also seeds the sandbox from the given environment
// spec. Each entry is either a bare NAME, forwarded from the host only when it
// is set, or an explicit NAME=value, set verbatim. The staging tier uses it to
// forward the MEGAPORT_* credentials and to pin MEGAPORT_ENVIRONMENT; hermetic
// tests do not.
func RunWithEnv(t *testing.T, envSpec []string, args ...string) Result {
	t.Helper()
	return run(t, envSpec, args)
}

func run(t *testing.T, envSpec []string, args []string) Result {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cliBinary, args...)
	cmd.Env = sandboxEnv(t, envSpec)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Check the deadline before unwrapping the exit code: a process killed on
	// timeout also returns an *exec.ExitError, but its code is meaningless. Gate
	// on err so a command that merely finished at the deadline is not misreported.
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("command timed out after %s: megaport-cli %s", runTimeout, strings.Join(args, " "))
	}

	exit := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exit = exitErr.ExitCode()
		} else {
			t.Fatalf("running megaport-cli %s: %v", strings.Join(args, " "), err)
		}
	}

	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Exit: exit}
}

// sandboxEnv builds the subprocess environment: an isolated HOME so the CLI
// never reads the developer's ~/.megaport/config.json, a minimal PATH, and the
// requested environment spec applied on top. Everything else, including all
// MEGAPORT_* config, is dropped so tests are hermetic by default.
func sandboxEnv(t *testing.T, envSpec []string) []string {
	t.Helper()
	// A fixed minimal PATH keeps the sandbox hermetic. The standard system dirs
	// cover the few tools the CLI may shell out to on its darwin and linux build
	// targets; git (used for the version string) is best-effort and degrades
	// gracefully when unresolved.
	env := []string{
		"HOME=" + t.TempDir(),
		"PATH=/usr/bin:/bin",
		// Suppress the GitHub update-check HTTP call so hermetic tests remain
		// network-free even when MEGAPORT_CLI_E2E_BIN points at a versioned binary.
		"NO_UPDATE_CHECK=1",
	}
	for _, entry := range envSpec {
		// An explicit NAME=value entry is set verbatim; a bare NAME is forwarded
		// from the host only when it is actually set.
		if strings.Contains(entry, "=") {
			env = append(env, entry)
			continue
		}
		if v, ok := os.LookupEnv(entry); ok {
			env = append(env, entry+"="+v)
		}
	}
	return env
}
