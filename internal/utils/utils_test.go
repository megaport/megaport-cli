package utils

import (
	"errors"
	"os"
	"testing"

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

	t.Run("invalid output format returns error", func(t *testing.T) {
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
