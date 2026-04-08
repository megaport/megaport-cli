package utils

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// LoginFunc is the function signature for authenticating with the Megaport API.
type LoginFunc func(ctx context.Context) (*megaport.Client, error)

// LoginClient creates a context from the command and logs in to the Megaport API.
// Returns the context, cancel function, and authenticated client.
// The caller must defer cancel().
func LoginClient(cmd *cobra.Command, defaultTimeout time.Duration, login LoginFunc) (context.Context, context.CancelFunc, *megaport.Client, error) {
	ctx, cancel := ContextFromCmdWithDefault(cmd, defaultTimeout)
	client, err := login(ctx)
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("error logging in: %w", err)
	}
	return ctx, cancel, client, nil
}
