package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
)

// RetryOpts configures the retry behaviour of RetryWithBackoff.
type RetryOpts struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

// DefaultRetryOpts returns sensible defaults for API retry behaviour.
// MaxRetries is taken from the --max-retries flag (default 3, set by cobra).
func DefaultRetryOpts() RetryOpts {
	return RetryOpts{
		MaxRetries:        MaxRetries,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// WithRetry wraps fn with the default retry policy. If --no-retry was set
// globally, fn is called exactly once.
func WithRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	if NoRetry {
		return fn(ctx)
	}
	return RetryWithBackoff(ctx, DefaultRetryOpts(), fn)
}

// RetryWithBackoff calls fn up to opts.MaxRetries+1 times with exponential
// backoff and jitter. It respects the Retry-After header from 429 responses
// and only retries on transient/server errors.
func RetryWithBackoff(ctx context.Context, opts RetryOpts, fn func(ctx context.Context) error) error {
	delay := opts.InitialDelay

	for attempt := 0; ; attempt++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}

		if attempt >= opts.MaxRetries {
			return err
		}

		if !isRetryable(err) {
			return err
		}

		// Use Retry-After header if present, otherwise exponential backoff.
		// Cap Retry-After at MaxDelay to prevent a misbehaving server from
		// stalling the CLI indefinitely.
		wait := retryAfterDelay(err)
		if wait > opts.MaxDelay {
			wait = opts.MaxDelay
		}
		if wait == 0 {
			wait = addJitter(delay)
		}
		// Ensure jittered wait also respects the hard cap.
		if wait > opts.MaxDelay {
			wait = opts.MaxDelay
		}

		logRetry(attempt+1, opts.MaxRetries, wait, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		// Increase delay for next attempt.
		delay = time.Duration(float64(delay) * opts.BackoffMultiplier)
		if delay > opts.MaxDelay {
			delay = opts.MaxDelay
		}
	}
}

// retryableStatusCodes lists HTTP status codes that warrant a retry.
var retryableStatusCodes = map[int]bool{
	429: true, // Too Many Requests
	502: true, // Bad Gateway
	503: true, // Service Unavailable
	504: true, // Gateway Timeout
}

// isRetryable returns true if err represents a transient failure.
func isRetryable(err error) bool {
	// Check for megaport API errors with retryable status codes.
	var apiErr *megaport.ErrorResponse
	if errors.As(err, &apiErr) && apiErr.Response != nil {
		return retryableStatusCodes[apiErr.Response.StatusCode]
	}

	// Check for transient network errors.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// Check for common transient error strings (connection reset, etc.).
	msg := err.Error()
	for _, s := range []string{
		"connection reset",
		"connection refused",
		"broken pipe",
		"EOF",
		"i/o timeout",
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}

// retryAfterDelay extracts the Retry-After header from a 429 response and
// returns the corresponding duration. Returns 0 if not applicable.
func retryAfterDelay(err error) time.Duration {
	var apiErr *megaport.ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Response == nil {
		return 0
	}
	if apiErr.Response.StatusCode != http.StatusTooManyRequests {
		return 0
	}
	ra := strings.TrimSpace(apiErr.Response.Header.Get("Retry-After"))
	if ra == "" {
		return 0
	}
	// Try parsing as seconds first (most common for APIs).
	if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	// Try parsing as HTTP-date.
	if t, err := http.ParseTime(ra); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// addJitter adds 10-25% random jitter to a duration.
func addJitter(d time.Duration) time.Duration {
	jitterFraction := 0.10 + rand.Float64()*0.15 // 10-25%
	jitter := time.Duration(float64(d) * jitterFraction)
	return d + jitter
}

// logRetry prints a retry message to stderr when verbose mode is active.
// attempt is 1-based (retry number), totalAttempts includes the initial call.
func logRetry(attempt, maxRetries int, wait time.Duration, err error) {
	totalAttempts := maxRetries + 1
	currentAttempt := attempt + 1 // +1 because attempt is retry index, display is overall attempt
	// Always log on the final retry attempt regardless of verbosity.
	if attempt == maxRetries {
		fmt.Fprintf(os.Stderr, "Retrying in %s (attempt %d/%d): %v\n", wait.Round(time.Millisecond), currentAttempt, totalAttempts, err)
		return
	}
	// For earlier attempts, only log in verbose mode.
	if isVerboseMode() {
		fmt.Fprintf(os.Stderr, "Retrying in %s (attempt %d/%d): %v\n", wait.Round(time.Millisecond), currentAttempt, totalAttempts, err)
	}
}

// isVerboseMode checks whether verbose mode is active. We duplicate a minimal
// check here rather than importing the output package, which would create a
// circular dependency (output -> utils -> output).
func isVerboseMode() bool {
	// The output package stores verbosity in an atomic.Value. We can't read
	// that from here, so we fall back to checking whether --verbose was set
	// via a package-level variable that the root command populates.
	return Verbose
}
