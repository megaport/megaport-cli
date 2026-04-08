package utils

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// WatchConfig holds configuration for the watch loop.
type WatchConfig struct {
	Interval     time.Duration
	NoColor      bool
	OutputFormat string
	ResourceType string // "Port", "VXC", "MCR", "MVE"
	ResourceUID  string
}

// WatchResource handles the common watch boilerplate: context creation, login,
// config setup, and WatchLoop invocation. The pollFn receives the authenticated
// client and should fetch the resource, print it, and return its provisioning status.
func WatchResource(
	cmd *cobra.Command,
	resourceType string,
	resourceUID string,
	noColor bool,
	outputFormat string,
	login LoginFunc,
	pollFn func(ctx context.Context, client *megaport.Client) (string, error),
) error {
	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return fmt.Errorf("invalid --interval flag: %w", err)
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be greater than 0")
	}

	ctx, cancel, client, err := LoginClient(cmd, DefaultWatchTimeout, login)
	if err != nil {
		return err
	}
	defer cancel()

	cfg := WatchConfig{
		Interval:     interval,
		NoColor:      noColor,
		OutputFormat: outputFormat,
		ResourceType: resourceType,
		ResourceUID:  resourceUID,
	}

	return WatchLoop(ctx, cfg, func(pollCtx context.Context) (string, error) {
		return pollFn(pollCtx, client)
	})
}
