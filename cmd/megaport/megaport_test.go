//go:build !js && !wasm
// +build !js,!wasm

package megaport

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noHeaderTestItem is a minimal struct used to observe table header suppression.
type noHeaderTestItem struct {
	Name string `header:"NAME" json:"name"`
}

// TestNoHeaderFlagWiredThroughPersistentPreRunE verifies that --no-header is
// registered on the root command and that PersistentPreRunE propagates its
// value to the output package by observing that table headers are suppressed.
func TestNoHeaderFlagWiredThroughPersistentPreRunE(t *testing.T) {
	// Restore all output state, the backing var, and the cobra flag so
	// subsequent tests in this package are not affected.
	defer func() {
		output.ResetState()
		noHeader = false
		_ = rootCmd.PersistentFlags().Set("no-header", "false")
	}()

	// Run a no-auth command through the real rootCmd so PersistentPreRunE fires.
	rootCmd.SetArgs([]string{"version", "--no-header"})
	_ = output.CaptureOutput(func() {
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	// Assert via observable behavior: table output should have no header row.
	captured := output.CaptureOutput(func() {
		_ = output.PrintOutput([]noHeaderTestItem{{Name: "row1"}}, "table", true)
	})
	assert.False(t, strings.Contains(captured, "NAME"), "header row should be suppressed after --no-header")
	assert.True(t, strings.Contains(captured, "row1"), "data rows should still appear")
}

// TestNoPagerDefaultApplied verifies that a "no-pager" default persisted in
// the config file is read by applyDefaultSettings AND forwarded to the output
// package via output.SetNoPager in PersistentPreRunE. Both sides of the
// wiring are asserted so that removing either call causes the test to fail.
func TestNoPagerDefaultApplied(t *testing.T) {
	// Use an isolated config directory so this test doesn't touch the real one.
	dir := t.TempDir()
	t.Setenv("MEGAPORT_CONFIG_DIR", dir)

	// Write no-pager = true as a config default.
	mgr, err := config.NewConfigManager()
	require.NoError(t, err)
	require.NoError(t, mgr.SetDefault("no-pager", true))

	// Restore package-level state and cobra flag after the test.
	defer func() {
		output.ResetState()
		noPager = false
		_ = rootCmd.PersistentFlags().Set("no-pager", "false")
	}()

	// Fire a lightweight command through the real rootCmd so PersistentPreRunE runs.
	rootCmd.SetArgs([]string{"version"})
	_ = output.CaptureOutput(func() {
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	// Assert the flag var was set (applyDefaultSettings path).
	assert.True(t, noPager, "noPager package var should be true after config default is applied")
	// Assert the output package was notified (output.SetNoPager wiring path).
	assert.True(t, output.GetNoPager(), "output.GetNoPager() should be true after PersistentPreRunE wires SetNoPager")
}

// TestApplyDefaultSettings_WarnsOnConfigLoadFailure verifies that when
// NewConfigManager fails (e.g. the configured config dir cannot be created),
// applyDefaultSettings emits a visible warning instead of returning silently.
func TestApplyDefaultSettings_WarnsOnConfigLoadFailure(t *testing.T) {
	// Create a temp file, then point MEGAPORT_CONFIG_DIR at a subpath of it.
	// os.MkdirAll will fail because the parent is a regular file, forcing
	// NewConfigManager to return an error.
	parent := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(parent, []byte("x"), 0600))
	t.Setenv("MEGAPORT_CONFIG_DIR", filepath.Join(parent, "child"))

	defer func() {
		output.ResetState()
		noColor = false
		_ = rootCmd.PersistentFlags().Set("no-color", "false")
	}()

	// PrintWarning writes to stdout in non-json output format; capture stdout.
	captured := output.CaptureOutput(func() {
		applyDefaultSettings(rootCmd)
	})
	assert.Contains(t, captured, "Could not load saved default settings")
}

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		// CLIError with each code
		{"CLIError general", exitcodes.New(exitcodes.General, errors.New("unknown")), exitcodes.General},
		{"CLIError usage", exitcodes.NewUsageError(errors.New("bad flag")), exitcodes.Usage},
		{"CLIError auth", exitcodes.NewAuthError(errors.New("no creds")), exitcodes.Authentication},
		{"CLIError api", exitcodes.NewAPIError(errors.New("500")), exitcodes.API},
		{"CLIError cancelled", exitcodes.NewCancelledError(errors.New("cancelled by user")), exitcodes.Cancelled},

		// Cobra-style errors
		{"cobra unknown command", errors.New(`unknown command "foo" for "megaport-cli"`), exitcodes.Usage},
		{"cobra unknown flag", errors.New(`unknown flag: --bogus`), exitcodes.Usage},
		{"cobra unknown shorthand flag", errors.New(`unknown shorthand flag: 'x' in -x`), exitcodes.Usage},
		{"cobra accepts at most", errors.New(`accepts at most 1 arg(s), received 2`), exitcodes.Usage},
		{"cobra required flag(s)", errors.New(`required flag(s) "name" not set`), exitcodes.Usage},

		// PersistentPreRunE format validation
		{"invalid output format", errors.New("invalid output format: yaml"), exitcodes.Usage},

		// Unknown errors
		{"unknown error", errors.New("something unexpected"), exitcodes.General},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, exitCodeFromError(tt.err))
		})
	}
}
