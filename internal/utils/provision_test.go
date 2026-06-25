package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestWaitForProvision_ReadyImmediately(t *testing.T) {
	calls := 0
	getStatus := func(ctx context.Context) (string, error) {
		calls++
		return megaport.SERVICE_LIVE, nil
	}

	err := WaitForProvision(context.Background(), "Port", "test-port", "uid-1", getStatus)

	assert.NoError(t, err)
	assert.Equal(t, 1, calls, "a ready resource should be detected on the first check without polling")
}

func TestWaitForProvision_ReturnsGetStatusError(t *testing.T) {
	sentinel := errors.New("boom")
	getStatus := func(ctx context.Context) (string, error) {
		return "", sentinel
	}

	err := WaitForProvision(context.Background(), "Port", "test-port", "uid-1", getStatus)

	assert.ErrorIs(t, err, sentinel)
}

func TestWaitForProvision_TerminalStateFailsFast(t *testing.T) {
	getStatus := func(ctx context.Context) (string, error) {
		return megaport.STATUS_CANCELLED, nil
	}

	err := WaitForProvision(context.Background(), "Port", "test-port", "uid-1", getStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terminal state")
}

func TestWaitForProvision_PollsUntilReady(t *testing.T) {
	orig := ProvisionPollInterval
	ProvisionPollInterval = time.Millisecond
	defer func() { ProvisionPollInterval = orig }()

	calls := 0
	getStatus := func(ctx context.Context) (string, error) {
		calls++
		if calls < 3 {
			return "DEPLOYING", nil
		}
		return megaport.SERVICE_LIVE, nil
	}

	err := WaitForProvision(context.Background(), "Port", "test-port", "uid-1", getStatus)

	assert.NoError(t, err)
	assert.Equal(t, 3, calls, "should poll until the resource reaches a ready state")
}

func TestWaitForProvision_TimesOut(t *testing.T) {
	orig := ProvisionPollInterval
	ProvisionPollInterval = time.Millisecond
	defer func() { ProvisionPollInterval = orig }()

	getStatus := func(ctx context.Context) (string, error) {
		return "DEPLOYING", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := WaitForProvision(ctx, "Port", "test-port", "uid-1", getStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
