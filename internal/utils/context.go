package utils

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// ContextFromCmd creates a context with timeout from the command's --timeout flag.
// If no timeout flag is set or the duration is zero, defaults to 90 seconds.
func ContextFromCmd(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return ContextFromCmdWithDefault(cmd, 90*time.Second)
}

// ContextFromCmdWithDefault creates a context with timeout from the command's
// --timeout flag, falling back to defaultTimeout when the flag is not set or
// is zero. Use this for long-running operations (e.g. provisioning) where the
// operation-appropriate default differs from the global 90-second default.
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
