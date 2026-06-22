package utils

import (
	"context"
	"fmt"
	"slices"
	"time"

	megaport "github.com/megaport/megaportgo"
)

// ProvisionPollInterval is how often WaitForProvision polls resource status.
// Overridable in tests to avoid real sleeps.
var ProvisionPollInterval = 10 * time.Second

// provisionReadyStates are the provisioning states considered fully provisioned.
var provisionReadyStates = []string{megaport.SERVICE_CONFIGURED, megaport.SERVICE_LIVE}

// provisionFailedStates are terminal states that mean provisioning will never
// succeed, so the wait loop should fail fast instead of polling until the timeout.
var provisionFailedStates = []string{megaport.STATUS_DECOMMISSIONED, megaport.STATUS_CANCELLED}

// WaitForProvision polls getStatus until the resource reaches a ready state,
// the caller's deadline elapses, ctx is cancelled, or getStatus returns an
// error. The order has already been placed by the time this runs, so it must
// never be wrapped in an order-submission retry.
func WaitForProvision(ctx context.Context, resType, name, uid string, getStatus func(ctx context.Context) (string, error)) error {
	check := func() (bool, error) {
		status, err := getStatus(ctx)
		if err != nil {
			return false, err
		}
		if slices.Contains(provisionFailedStates, status) {
			return false, fmt.Errorf("%s %q (%s) entered terminal state %q during provisioning", resType, name, uid, status)
		}
		return slices.Contains(provisionReadyStates, status), nil
	}

	if ready, err := check(); err != nil || ready {
		return err
	}

	// Respect the caller's deadline (e.g. from --timeout); only impose the
	// default cap when the caller passed an open-ended context.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultProvisionTimeout)
		defer cancel()
	}

	ticker := time.NewTicker(ProvisionPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for %s %q (%s) to provision: %w", resType, name, uid, ctx.Err())
		case <-ticker.C:
			if ready, err := check(); err != nil || ready {
				return err
			}
		}
	}
}
