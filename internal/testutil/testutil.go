package testutil

import (
	"context"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// NoColorAdapter wraps a function that takes (cmd, args, noColor) into a
// standard cobra RunE, always passing noColor=true for tests.
func NoColorAdapter(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(cmd, args, true)
	}
}

// OutputAdapter wraps a function that takes (cmd, args, noColor, outputFormat)
// into a standard cobra RunE. It reads --output from the command flags,
// defaulting to "table".
func OutputAdapter(fn func(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("output")
		if format == "" {
			format = "table"
		}
		return fn(cmd, args, true, format)
	}
}

// SetupLogin saves the current config.LoginFunc, replaces it with one that
// returns a client configured by setupFn, and returns a cleanup function
// that restores the original.
func SetupLogin(setupFn func(*megaport.Client)) func() {
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		setupFn(client)
		return client, nil
	})
	return func() {
		config.SetLoginFunc(original)
	}
}

// SetupLoginError saves the current config.LoginFunc, replaces it with one
// that returns the given error, and returns a cleanup function.
func SetupLoginError(err error) func() {
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return nil, err
	})
	return func() {
		config.SetLoginFunc(original)
	}
}

// NewCommand creates a cobra.Command with the given use string and RunE,
// pre-configured with the --output flag defaulting to "table".
func NewCommand(use string, runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:  use,
		RunE: runE,
	}
	cmd.Flags().StringP("output", "o", "table", "Output format")
	return cmd
}

// SetFlags sets multiple string flag values on a command, calling t.Fatalf
// on any error.
func SetFlags(t *testing.T, cmd *cobra.Command, flags map[string]string) {
	t.Helper()
	for k, v := range flags {
		if err := cmd.Flags().Set(k, v); err != nil {
			t.Fatalf("Failed to set flag %q to %q: %v", k, v, err)
		}
	}
}
