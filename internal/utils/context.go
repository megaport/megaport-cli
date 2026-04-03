package utils

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// ContextFromCmd creates a context with timeout from the command's --timeout flag.
// If no timeout flag is set or the duration is zero, defaults to 90 seconds.
func ContextFromCmd(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	timeout := 90 * time.Second

	// Try to read the timeout flag
	if cmd != nil && cmd.Flags() != nil {
		if val, err := cmd.Flags().GetDuration("timeout"); err == nil && val > 0 {
			timeout = val
		}
	}

	return context.WithTimeout(context.Background(), timeout)
}
