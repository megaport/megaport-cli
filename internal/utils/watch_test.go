package utils

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
