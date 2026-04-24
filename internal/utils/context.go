package utils

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

const (
	// DefaultProvisionTimeout is the default wait time for provisioning operations
	// (buy, update commands that wait for the resource to become ready).
	DefaultProvisionTimeout = 10 * time.Minute

	// DefaultMutationTimeout is the default context timeout for buy/update/delete
	// commands that may take longer than the standard request timeout.
	DefaultMutationTimeout = 15 * time.Minute

	// DefaultWatchTimeout is the default timeout for watch mode loops.
	// Watch mode monitors resource status which can take 15+ minutes during
	// provisioning, so this is set generously. Users can override with --timeout.
	DefaultWatchTimeout = 30 * time.Minute
)

// ContextFromCmd creates a context with timeout from the command's --timeout flag.
// If no timeout flag is set or the duration is zero or negative, defaults to 90 seconds.
func ContextFromCmd(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return ContextFromCmdWithDefault(cmd, 90*time.Second)
}

// ContextFromCmdWithDefault creates a context with timeout from the command's
// --timeout flag, falling back to defaultTimeout when the flag is not set or
// is zero or negative. Use this for long-running operations (e.g. provisioning)
// where the operation-appropriate default differs from the global 90-second default.
func ContextFromCmdWithDefault(cmd *cobra.Command, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	timeout := defaultTimeout

	// Try to read the timeout flag; user-supplied value always wins.
	if cmd != nil {
		if val, err := cmd.Flags().GetDuration("timeout"); err == nil && val > 0 {
			timeout = val
		}
	}

	return context.WithTimeout(context.Background(), timeout)
}
