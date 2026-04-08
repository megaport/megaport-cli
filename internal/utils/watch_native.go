//go:build !wasm
// +build !wasm

package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
)

// WatchLoop runs pollFn repeatedly at the configured interval until interrupted.
// pollFn should fetch and print its output, returning the current status string
// for transition tracking.
func WatchLoop(ctx context.Context, cfg WatchConfig, pollFn func(ctx context.Context) (string, error)) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	isTable := cfg.OutputFormat == "" || cfg.OutputFormat == "table"
	var lastStatus string

	// Run immediately on first call, then on each tick
	for first := true; ; first = false {
		if !first {
			select {
			case <-ctx.Done():
				fmt.Println()
				if ctx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("watch timed out for %s %s — use --timeout <duration> to extend (e.g., --timeout 1h): %w",
						cfg.ResourceType, cfg.ResourceUID, ctx.Err())
				}
				output.PrintInfo("Watch stopped.", cfg.NoColor)
				return nil
			case <-ticker.C:
			}
		}

		if isTable {
			output.ClearScreen()
			fmt.Printf("Every %s: %s %s    %s\n\n", cfg.Interval, cfg.ResourceType, cfg.ResourceUID, time.Now().Format("2006-01-02 15:04:05"))
		} else {
			if !first {
				fmt.Println("---")
			}
		}

		currentStatus, err := pollFn(ctx)
		if err != nil {
			output.PrintError("Error polling %s %s: %v", cfg.NoColor, cfg.ResourceType, cfg.ResourceUID, err)
		} else if lastStatus != "" && currentStatus != lastStatus {
			fmt.Println()
			output.PrintSuccess("Status changed: %s → %s", cfg.NoColor, lastStatus, currentStatus)
		}

		if currentStatus != "" {
			lastStatus = currentStatus
		}

		fmt.Printf("\nLast checked: %s\n", time.Now().Format(time.RFC3339))
	}
}
