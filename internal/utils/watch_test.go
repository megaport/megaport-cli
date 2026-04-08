//go:build !wasm
// +build !wasm

package utils

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// --- WatchLoop tests (existing) ---

func TestWatchLoop_PollsAndDetectsTransition(t *testing.T) {
	var callCount atomic.Int32
	statuses := []string{"CONFIGURED", "CONFIGURED", "LIVE"}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := WatchConfig{
		Interval:     50 * time.Millisecond,
		NoColor:      true,
		OutputFormat: "json",
		ResourceType: "Port",
		ResourceUID:  "test-uid",
	}

	pollFn := func(_ context.Context) (string, error) {
		idx := int(callCount.Add(1)) - 1
		if idx >= len(statuses) {
			cancel()
			return statuses[len(statuses)-1], nil
		}
		return statuses[idx], nil
	}

	err := WatchLoop(ctx, cfg, pollFn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int(callCount.Load()), 3)
}

func TestWatchLoop_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := WatchConfig{
		Interval:     50 * time.Millisecond,
		NoColor:      true,
		OutputFormat: "json",
		ResourceType: "Port",
		ResourceUID:  "test-uid",
	}

	var callCount atomic.Int32
	pollFn := func(_ context.Context) (string, error) {
		if callCount.Add(1) >= 2 {
			cancel()
		}
		return "CONFIGURED", nil
	}

	err := WatchLoop(ctx, cfg, pollFn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int(callCount.Load()), 2)
}

func TestWatchLoop_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	cfg := WatchConfig{
		Interval:     50 * time.Millisecond,
		NoColor:      true,
		OutputFormat: "json",
		ResourceType: "Port",
		ResourceUID:  "test-uid",
	}

	var callCount atomic.Int32
	pollFn := func(_ context.Context) (string, error) {
		callCount.Add(1)
		return "CONFIGURED", nil
	}

	err := WatchLoop(ctx, cfg, pollFn)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, int(callCount.Load()), 1, "should poll at least once before timeout")
}

func TestWatchLoop_PollError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := WatchConfig{
		Interval:     50 * time.Millisecond,
		NoColor:      true,
		OutputFormat: "json",
		ResourceType: "Port",
		ResourceUID:  "test-uid",
	}

	var callCount atomic.Int32
	pollFn := func(_ context.Context) (string, error) {
		count := callCount.Add(1)
		if count == 1 {
			return "", fmt.Errorf("api error")
		}
		cancel()
		return "LIVE", nil
	}

	err := WatchLoop(ctx, cfg, pollFn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int(callCount.Load()), 2)
}

// --- WatchResource tests ---

// newWatchTestCmd creates a cobra command with the flags that WatchResource reads.
func newWatchTestCmd(interval time.Duration) *cobra.Command {
	cmd := &cobra.Command{Use: "watch-test"}
	cmd.Flags().Duration("interval", interval, "poll interval")
	cmd.Flags().Duration("timeout", 0, "timeout")
	return cmd
}

func TestWatchResource(t *testing.T) {
	tests := []struct {
		name        string
		interval    time.Duration
		login       LoginFunc
		wantErr     bool
		errContains string
	}{
		{
			name:     "invalid interval zero returns error",
			interval: 0,
			login: func(_ context.Context) (*megaport.Client, error) {
				return &megaport.Client{}, nil
			},
			wantErr:     true,
			errContains: "must be greater than 0",
		},
		{
			name:     "negative interval returns error",
			interval: -5 * time.Second,
			login: func(_ context.Context) (*megaport.Client, error) {
				return &megaport.Client{}, nil
			},
			wantErr:     true,
			errContains: "must be greater than 0",
		},
		{
			name:     "login failure returns wrapped error",
			interval: 5 * time.Second,
			login: func(_ context.Context) (*megaport.Client, error) {
				return nil, fmt.Errorf("auth failed")
			},
			wantErr:     true,
			errContains: "error logging in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newWatchTestCmd(tt.interval)

			// pollFn should never be called for these error cases
			pollFn := func(_ context.Context, _ *megaport.Client) (string, error) {
				t.Fatal("pollFn should not be called")
				return "", nil
			}

			err := WatchResource(cmd, "Port", "test-uid-123", true, "json", tt.login, pollFn)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWatchResource_IntervalPropagation(t *testing.T) {
	// Verify that a valid interval + successful login reaches WatchLoop (which
	// will tick at the configured interval). We use a very short interval and
	// a context timeout to let WatchLoop run briefly, then confirm polling
	// actually happened — proving the interval was propagated.

	mockClient := &megaport.Client{}
	login := func(_ context.Context) (*megaport.Client, error) {
		return mockClient, nil
	}

	cmd := newWatchTestCmd(50 * time.Millisecond)
	// Set a short timeout so WatchLoop exits quickly via deadline.
	assert.NoError(t, cmd.Flags().Set("timeout", "300ms"))

	var pollCount atomic.Int32
	pollFn := func(_ context.Context, client *megaport.Client) (string, error) {
		assert.Equal(t, mockClient, client)
		pollCount.Add(1)
		return "CONFIGURED", nil
	}

	err := WatchResource(cmd, "Port", "test-uid-456", true, "json", login, pollFn)
	// WatchLoop returns a timeout error when the context deadline exceeds.
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	// With a 50ms interval and 300ms timeout, we expect at least 2 polls.
	assert.GreaterOrEqual(t, int(pollCount.Load()), 2, "should have polled multiple times, confirming interval was used")
}
