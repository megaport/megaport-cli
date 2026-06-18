//go:build e2e

// Package e2e holds native-binary black-box tests. They build the CLI once and
// drive it via argv, exercising argv parsing, exit codes, and stdout/stderr in a
// way the in-process unit and integration tests cannot. Everything here is behind
// the e2e build tag, so a plain `go test ./...` and the production build ignore it.
package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// cliBinary is the absolute path to the CLI under test, resolved once by TestMain.
var cliBinary string

func TestMain(m *testing.M) {
	code, err := setupAndRun(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e setup failed:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// setupAndRun resolves the binary (building it unless MEGAPORT_CLI_E2E_BIN points
// at a prebuilt one) and then runs the suite.
func setupAndRun(m *testing.M) (int, error) {
	if prebuilt := os.Getenv("MEGAPORT_CLI_E2E_BIN"); prebuilt != "" {
		abs, err := filepath.Abs(prebuilt)
		if err != nil {
			return 0, fmt.Errorf("resolving MEGAPORT_CLI_E2E_BIN %q: %w", prebuilt, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return 0, fmt.Errorf("MEGAPORT_CLI_E2E_BIN %q: %w", abs, err)
		}
		if !info.Mode().IsRegular() {
			return 0, fmt.Errorf("MEGAPORT_CLI_E2E_BIN %q is not a regular file", abs)
		}
		cliBinary = abs
		return m.Run(), nil
	}

	root, err := moduleRoot()
	if err != nil {
		return 0, err
	}

	tmp, err := os.MkdirTemp("", "megaport-cli-e2e")
	if err != nil {
		return 0, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "megaport-cli")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	// Build from the module's own definition, ignoring any ambient go.work, so the
	// e2e binary is the same one a standalone or CI checkout produces.
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("building CLI binary: %w\n%s", err, out)
	}
	cliBinary = bin

	return m.Run(), nil
}

// moduleRoot returns the module root, which is the parent of the e2e package.
// `go test` sets the working directory to the package source directory. The
// go.mod check guards against the package being moved deeper, which would make
// the parent dir the wrong build target.
func moduleRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolving working directory: %w", err)
	}
	root := filepath.Dir(cwd)
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return "", fmt.Errorf("no go.mod at expected module root %q: %w", root, err)
	}
	return root, nil
}
