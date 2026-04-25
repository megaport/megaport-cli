package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// Verbosity check functions, defaulting to the output package. Tests can
// replace these to avoid mutating global verbosity state.
var (
	isQuietFunc   = output.IsQuiet
	isVerboseFunc = output.IsVerbose
)

// RetryOpts configures the retry behaviour of RetryWithBackoff.
type RetryOpts struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	// RetryNetworkErrors enables retrying on ambiguous network failures
	// (connection reset, EOF, timeout). Only set this for idempotent operations.
	RetryNetworkErrors bool
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

// WithRetry wraps fn with the default retry policy, only retrying on
// server-confirmed errors (HTTP status codes like 429, 502, 503, 504).
// Network-level errors (connection reset, EOF, timeout) are NOT retried
// because it is unsafe for non-idempotent operations — the server may
// have processed the request despite the client-side failure.
// Use WithIdempotentRetry for read-only or otherwise idempotent operations
// where retrying network errors is safe.
// If --no-retry was set globally, fn is called exactly once.
func WithRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	if NoRetry {
		return fn(ctx)
	}
	return RetryWithBackoff(ctx, DefaultRetryOpts(), fn)
}

// WithIdempotentRetry wraps fn with the default retry policy including
// network-level error retries. Use this for read-only or idempotent
// operations where retrying after an ambiguous network failure is safe.
// If --no-retry was set globally, fn is called exactly once.
func WithIdempotentRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	if NoRetry {
		return fn(ctx)
	}
	opts := DefaultRetryOpts()
	opts.RetryNetworkErrors = true
	return RetryWithBackoff(ctx, opts, fn)
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

		if !isRetryable(err, opts.RetryNetworkErrors) {
			return err
		}

		// Use Retry-After header if present, otherwise exponential backoff.
		// Cap Retry-After at MaxDelay to prevent a misbehaving server from
		// stalling the CLI indefinitely.
		wait, hasRetryAfter := retryAfterDelay(err)
		if wait > opts.MaxDelay {
			wait = opts.MaxDelay
		}
		if !hasRetryAfter {
			wait = addJitter(delay)
		}
		// Ensure jittered wait also respects the hard cap.
		if wait > opts.MaxDelay {
			wait = opts.MaxDelay
		}

		logRetry(attempt+1, opts.MaxRetries, wait, err)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
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
// When retryNetworkErrors is false, only server-confirmed errors (HTTP status
// codes) are considered retryable — ambiguous network failures are not.
func isRetryable(err error, retryNetworkErrors bool) bool {
	// Check for megaport API errors with retryable status codes.
	var apiErr *megaport.ErrorResponse
	if errors.As(err, &apiErr) && apiErr.Response != nil {
		return retryableStatusCodes[apiErr.Response.StatusCode]
	}

	// Network-level errors are only retried for idempotent operations.
	if !retryNetworkErrors {
		return false
	}

	// Structured sentinel/type checks first.
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// syscall-level transient errors (ECONNRESET, ECONNREFUSED, EPIPE).
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}

	// Check for transient network errors via the net.Error interface.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// Last-resort substring matching for errors that don't expose structured types.
	msg := strings.ToLower(err.Error())
	for _, s := range []string{
		"connection reset",
		"connection refused",
		"broken pipe",
		"i/o timeout",
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}

// retryAfterDelay extracts the Retry-After header from a 429 response and
// returns the corresponding duration and whether a valid header was found.
// A Retry-After value of "0" is valid and means retry immediately.
func retryAfterDelay(err error) (time.Duration, bool) {
	var apiErr *megaport.ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Response == nil {
		return 0, false
	}
	if apiErr.Response.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}
	ra := strings.TrimSpace(apiErr.Response.Header.Get("Retry-After"))
	if ra == "" {
		return 0, false
	}
	// Try parsing as seconds first (most common for APIs).
	if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second, true
	}
	// Try parsing as HTTP-date.
	if t, err := http.ParseTime(ra); err == nil {
		if d := time.Until(t); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// addJitter adds 10-25% random jitter to a duration.
func addJitter(d time.Duration) time.Duration {
	jitterFraction := 0.10 + rand.Float64()*0.15 // 10-25%
	jitter := time.Duration(float64(d) * jitterFraction)
	return d + jitter
}

// logRetry prints a retry message to stderr.
// attempt is the 1-based retry number; the displayed attempt count includes the initial call.
func logRetry(attempt, maxRetries int, wait time.Duration, err error) {
	// Suppress all retry logs in quiet mode.
	if isQuietFunc() {
		return
	}
	totalAttempts := maxRetries + 1
	overallAttempt := attempt + 1 // retry #1 happens after the initial call, so it is overall attempt #2
	// Always log on the final retry attempt regardless of verbosity.
	if attempt == maxRetries {
		fmt.Fprintf(os.Stderr, "Retrying in %s (attempt %d/%d): %v\n", wait.Round(time.Millisecond), overallAttempt, totalAttempts, err)
		return
	}
	// For earlier attempts, only log in verbose mode.
	if isVerboseFunc() {
		fmt.Fprintf(os.Stderr, "Retrying in %s (attempt %d/%d): %v\n", wait.Round(time.Millisecond), overallAttempt, totalAttempts, err)
	}
}
