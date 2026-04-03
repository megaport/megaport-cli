package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldDisableColors(t *testing.T) {
	// Save original state and restore after each subtest.
	origNoColor := NoColor
	origArgs := os.Args
	origNoColorEnv, origNoColorEnvSet := os.LookupEnv("NO_COLOR")
	defer func() {
		NoColor = origNoColor
		os.Args = origArgs
		if origNoColorEnvSet {
			os.Setenv("NO_COLOR", origNoColorEnv)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	t.Run("returns true when NO_COLOR env is set", func(t *testing.T) {
		NoColor = false
		os.Args = origArgs
		t.Setenv("NO_COLOR", "1")
		assert.True(t, ShouldDisableColors())
	})

	t.Run("returns true when NoColor global var is set", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Args = []string{"cmd"}
		NoColor = true
		assert.True(t, ShouldDisableColors())
		NoColor = false
	})

	t.Run("returns true when --no-color arg is present", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		NoColor = false
		os.Args = []string{"cmd", "--no-color"}
		assert.True(t, ShouldDisableColors())
	})

	t.Run("returns false when neither is set", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		NoColor = false
		os.Args = []string{"cmd"}
		assert.False(t, ShouldDisableColors())
	})
}

func TestGetCurrentEnv(t *testing.T) {
	origEnv := Env
	defer func() { Env = origEnv }()

	t.Run("returns production by default", func(t *testing.T) {
		Env = ""
		assert.Equal(t, "production", GetCurrentEnv())
	})

	t.Run("returns configured env if set", func(t *testing.T) {
		Env = "staging"
		assert.Equal(t, "staging", GetCurrentEnv())
	})
}

func TestWrapRunE(t *testing.T) {
	t.Run("success returns nil", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
			return nil
		})
		cmd := &cobra.Command{Use: "test"}
		err := wrapped(cmd, []string{})
		assert.NoError(t, err)
		// SilenceUsage should not be set on success.
		assert.False(t, cmd.SilenceUsage)
	})

	t.Run("error returns formatted error and silences usage", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
			return errors.New("something went wrong")
		})
		cmd := &cobra.Command{Use: "test"}
		err := wrapped(cmd, []string{"arg1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error running test command")
		assert.Contains(t, err.Error(), "something went wrong")
		assert.Contains(t, err.Error(), "arg1")
		assert.True(t, cmd.SilenceUsage)
		assert.True(t, cmd.SilenceErrors)

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.General, cliErr.Code)
	})

	t.Run("error message includes command name and args", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
			return errors.New("fail")
		})
		cmd := &cobra.Command{Use: "deploy"}
		err := wrapped(cmd, []string{"a", "b"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Command: deploy")
		assert.Contains(t, err.Error(), "[a b]")
	})

	t.Run("query flag without json output returns usage error", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
			return nil
		})
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().String("query", "", "")
		root.PersistentFlags().String("fields", "", "")
		root.PersistentFlags().String("output", "table", "")
		child := &cobra.Command{Use: "version"}
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("query", "[*].uid"))

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--query flag requires --output json")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("query flag with json output passes format guard", func(t *testing.T) {
		called := false
		wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
			called = true
			return nil
		})
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().String("query", "", "")
		root.PersistentFlags().String("fields", "", "")
		root.PersistentFlags().String("output", "table", "")
		child := &cobra.Command{Use: "version"}
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("query", "[*].uid"))
		require.NoError(t, root.PersistentFlags().Set("output", "json"))

		err := wrapped(child, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWrapColorAwareRunE(t *testing.T) {
	t.Run("extracts noColor flag correctly when true", func(t *testing.T) {
		var capturedNoColor bool
		wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
			capturedNoColor = noColor
			return nil
		})

		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		child := &cobra.Command{Use: "child"}
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("no-color", "true"))

		err := wrapped(child, []string{})
		assert.NoError(t, err)
		assert.True(t, capturedNoColor)
	})

	t.Run("defaults noColor to false when flag missing", func(t *testing.T) {
		var capturedNoColor bool
		wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
			capturedNoColor = noColor
			return nil
		})

		// Command without the no-color flag on root.
		cmd := &cobra.Command{Use: "orphan"}
		err := wrapped(cmd, []string{})
		assert.NoError(t, err)
		assert.False(t, capturedNoColor)
	})

	t.Run("error is formatted", func(t *testing.T) {
		wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
			return errors.New("color error")
		})
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		child := &cobra.Command{Use: "child"}
		root.AddCommand(child)

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error running child command")
		assert.Contains(t, err.Error(), "color error")
		assert.True(t, child.SilenceUsage)
	})

	t.Run("query flag without json output returns usage error", func(t *testing.T) {
		wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
			return nil
		})
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		root.PersistentFlags().String("query", "", "")
		root.PersistentFlags().String("fields", "", "")
		root.PersistentFlags().String("output", "table", "")
		child := &cobra.Command{Use: "status"}
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("query", "[*].uid"))

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--query flag requires --output json")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("query flag with json output passes format guard", func(t *testing.T) {
		called := false
		wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
			called = true
			return nil
		})
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		root.PersistentFlags().String("query", "", "")
		root.PersistentFlags().String("fields", "", "")
		root.PersistentFlags().String("output", "table", "")
		child := &cobra.Command{Use: "status"}
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("query", "[*].uid"))
		require.NoError(t, root.PersistentFlags().Set("output", "json"))

		err := wrapped(child, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWrapOutputFormatRunE(t *testing.T) {
	t.Run("extracts output format and noColor correctly", func(t *testing.T) {
		var capturedFormat string
		var capturedNoColor bool
		wrapped := WrapOutputFormatRunE(func(cmd *cobra.Command, args []string, noColor bool, format string) error {
			capturedFormat = format
			capturedNoColor = noColor
			return nil
		})

		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		child := &cobra.Command{Use: "list"}
		child.Flags().String("output", "json", "")
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("no-color", "true"))

		err := wrapped(child, []string{})
		assert.NoError(t, err)
		assert.Equal(t, "json", capturedFormat)
		assert.True(t, capturedNoColor)
	})

	t.Run("defaults to table format when flag missing", func(t *testing.T) {
		var capturedFormat string
		wrapped := WrapOutputFormatRunE(func(cmd *cobra.Command, args []string, noColor bool, format string) error {
			capturedFormat = format
			return nil
		})

		// No output flag defined at all.
		cmd := &cobra.Command{Use: "test"}
		err := wrapped(cmd, []string{})
		assert.NoError(t, err)
		assert.Equal(t, FormatTable, capturedFormat)
	})

	t.Run("invalid output format returns error with usage exit code", func(t *testing.T) {
		wrapped := WrapOutputFormatRunE(func(cmd *cobra.Command, args []string, noColor bool, format string) error {
			return nil
		})

		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		child := &cobra.Command{Use: "list"}
		child.Flags().String("output", "", "")
		root.AddCommand(child)
		require.NoError(t, child.Flags().Set("output", "yaml"))

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output format: yaml")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("query flag with non-json format returns usage error", func(t *testing.T) {
		wrapped := WrapOutputFormatRunE(func(cmd *cobra.Command, args []string, noColor bool, format string) error {
			return nil
		})

		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		root.PersistentFlags().String("query", "", "")
		root.PersistentFlags().String("fields", "", "")
		// --output is registered as a local flag on the child to match production
		// usage (WrapOutputFormatRunE reads it via cmd.Flags(), not PersistentFlags).
		child := &cobra.Command{Use: "list"}
		child.Flags().String("output", "table", "")
		root.AddCommand(child)
		require.NoError(t, root.PersistentFlags().Set("query", "[*].uid"))
		// output is "table" (the default set above); guard should reject it.

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--query flag requires --output json")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("function error is formatted", func(t *testing.T) {
		wrapped := WrapOutputFormatRunE(func(cmd *cobra.Command, args []string, noColor bool, format string) error {
			return errors.New("inner failure")
		})

		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool("no-color", false, "")
		child := &cobra.Command{Use: "get"}
		child.Flags().String("output", "table", "")
		root.AddCommand(child)

		err := wrapped(child, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error running get command")
		assert.Contains(t, err.Error(), "inner failure")
	})
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		wantCode int
	}{
		// Auth patterns
		{"auth - error logging in", "error logging in: bad creds", exitcodes.Authentication},
		{"auth - access key not provided", "access key not provided", exitcodes.Authentication},
		{"auth - secret key not provided", "secret key not provided", exitcodes.Authentication},
		{"auth - authentication failed", "authentication failed for user", exitcodes.Authentication},
		{"auth - Authorize", "failed to Authorize request", exitcodes.Authentication},

		// Usage patterns
		{"usage - invalid output format", "invalid output format: yaml", exitcodes.Usage},
		{"usage - required flag", "required flag \"name\" not set", exitcodes.Usage},
		{"usage - not set when not using interactive", "required flag \"term\" not set when not using interactive or JSON input", exitcodes.Usage},
		{"usage - at least one field", "at least one field must be updated", exitcodes.Usage},
		{"usage - at least one of these flags", "at least one of these flags must be set: name, term", exitcodes.Usage},
		{"usage - invalid location ID", "invalid location ID: abc", exitcodes.Usage},
		{"usage - invalid ID combo", "invalid port ID provided", exitcodes.Usage},

		// API patterns
		{"api - error listing", "error listing ports", exitcodes.API},
		{"api - error getting", "error getting VXC details", exitcodes.API},
		{"api - error creating", "error creating MCR", exitcodes.API},
		{"api - error updating", "error updating port name", exitcodes.API},
		{"api - error deleting", "error deleting service", exitcodes.API},
		{"api - error buying", "error buying port", exitcodes.API},
		{"api - error modifying", "error modifying VXC", exitcodes.API},
		{"api - failed to retrieve", "failed to retrieve port info", exitcodes.API},
		{"api - failed to buy", "failed to buy VXC", exitcodes.API},
		{"api - failed to validate", "failed to validate order", exitcodes.API},
		{"api - API failure", "API failure: 500 internal server error", exitcodes.API},

		// General/unknown
		{"general - unknown error", "something unexpected happened", exitcodes.General},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := classifyError(errors.New(tt.errMsg))
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestClassifyError_CLIErrorPassthrough(t *testing.T) {
	// A CLIError already carrying a code should have that code preserved.
	inner := exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
	assert.Equal(t, exitcodes.Cancelled, classifyError(inner))

	// Works through wrapping too.
	wrapped := fmt.Errorf("outer: %w", inner)
	assert.Equal(t, exitcodes.Cancelled, classifyError(wrapped))
}

func TestClassifyError_SDKErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   int
	}{
		{"401 -> Authentication", 401, exitcodes.Authentication},
		{"403 -> Authentication", 403, exitcodes.Authentication},
		{"404 -> API", 404, exitcodes.API},
		{"422 -> API", 422, exitcodes.API},
		{"429 -> API", 429, exitcodes.API},
		{"500 -> API", 500, exitcodes.API},
		{"502 -> API", 502, exitcodes.API},
		{"503 -> API", 503, exitcodes.API},
		{"200 -> General (no match)", 200, exitcodes.General},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := classifyError(makeAPIError(tt.statusCode, ""))
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestWrapRunE_CancelledError(t *testing.T) {
	// WrapRunE must preserve the Cancelled code when the action function
	// returns a CLIError that already carries it.
	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		return exitcodes.NewCancelledError(fmt.Errorf("cancelled by user"))
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Cancelled, cliErr.Code)
}

func TestWrapColorAwareRunE_CancelledError(t *testing.T) {
	wrapped := WrapColorAwareRunE(func(cmd *cobra.Command, args []string, noColor bool) error {
		return exitcodes.NewCancelledError(fmt.Errorf("cancelled by user"))
	})
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("no-color", false, "")
	err := wrapped(cmd, []string{})
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Cancelled, cliErr.Code)
}

func TestWrapRunE_AuthError(t *testing.T) {
	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		return errors.New("error logging in: invalid credentials")
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Authentication, cliErr.Code)
}

func TestWrapRunE_APIError(t *testing.T) {
	wrapped := WrapRunE(func(cmd *cobra.Command, args []string) error {
		return errors.New("error listing ports: connection refused")
	})
	cmd := &cobra.Command{Use: "test"}
	err := wrapped(cmd, []string{})
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.API, cliErr.Code)
}
