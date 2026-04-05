package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: create a megaport ErrorResponse with the given status code and optional Retry-After header.
func apiError(statusCode int, retryAfter string) error {
	header := http.Header{}
	if retryAfter != "" {
		header.Set("Retry-After", retryAfter)
	}
	return &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: statusCode,
			Header:     header,
			Request:    &http.Request{URL: &url.URL{}},
		},
		Message: fmt.Sprintf("status %d", statusCode),
	}
}

func TestRetryWithBackoff_SuccessFirstAttempt(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryWithBackoff_SuccessAfterTransientFailure(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return apiError(503, "")
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
		calls++
		return apiError(503, "")
	})
	assert.Error(t, err)
	assert.Equal(t, 4, calls) // 1 initial + 3 retries
}

func TestRetryWithBackoff_NonRetryableErrorFailsImmediately(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", 400},
		{"401 Unauthorized", 401},
		{"403 Forbidden", 403},
		{"404 Not Found", 404},
		{"409 Conflict", 409},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			err := RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
				calls++
				return apiError(tt.statusCode, "")
			})
			assert.Error(t, err)
			assert.Equal(t, 1, calls, "non-retryable error should not trigger retry")
		})
	}
}

func TestRetryWithBackoff_RetryableStatusCodes(t *testing.T) {
	for _, code := range []int{429, 502, 503, 504} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			calls := 0
			_ = RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
				calls++
				return apiError(code, "")
			})
			assert.Equalf(t, 4, calls, "retryable status %d should exhaust all retries", code)
		})
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := RetryWithBackoff(ctx, RetryOpts{
		MaxRetries:        10,
		InitialDelay:      50 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 1.0,
	}, func(ctx context.Context) error {
		calls++
		return apiError(503, "")
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Less(t, calls, 10, "should not exhaust all retries when context cancelled")
}

func TestRetryWithBackoff_RetryAfterHeaderRespected(t *testing.T) {
	calls := 0
	start := time.Now()
	err := RetryWithBackoff(context.Background(), RetryOpts{
		MaxRetries:        1,
		InitialDelay:      10 * time.Second, // very long default — should be overridden by Retry-After
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}, func(ctx context.Context) error {
		calls++
		if calls == 1 {
			return apiError(429, "1") // Retry-After: 1 second
		}
		return nil
	})
	elapsed := time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, 2, calls)
	// Should have waited ~1s (from Retry-After), not ~10s (from InitialDelay)
	assert.Less(t, elapsed, 5*time.Second, "should use Retry-After delay, not InitialDelay")
}

func TestRetryWithBackoff_TransientNetworkErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"timeout error", &net.DNSError{IsTimeout: true}, true},
		{"connection reset (string)", fmt.Errorf("read tcp: connection reset by peer"), true},
		{"connection refused (string)", fmt.Errorf("dial tcp: connection refused"), true},
		{"io.EOF", io.EOF, true},
		{"io.ErrUnexpectedEOF", io.ErrUnexpectedEOF, true},
		{"wrapped io.EOF", fmt.Errorf("read failed: %w", io.EOF), true},
		{"i/o timeout (string)", fmt.Errorf("i/o timeout"), true},
		{"context.DeadlineExceeded", context.DeadlineExceeded, true},
		{"wrapped DeadlineExceeded", fmt.Errorf("request failed: %w", context.DeadlineExceeded), true},
		{"syscall ECONNRESET", fmt.Errorf("write: %w", syscall.ECONNRESET), true},
		{"syscall ECONNREFUSED", fmt.Errorf("dial: %w", syscall.ECONNREFUSED), true},
		{"syscall EPIPE", fmt.Errorf("write: %w", syscall.EPIPE), true},
		{"plain error", fmt.Errorf("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Network errors are retryable only when retryNetworkErrors is true (idempotent operations).
			assert.Equal(t, tt.retryable, isRetryable(tt.err, true))
		})
	}
}

func TestIsRetryable_MutationMode(t *testing.T) {
	// With retryNetworkErrors=false, network errors should NOT be retryable.
	assert.False(t, isRetryable(&net.DNSError{IsTimeout: true}, false), "network timeout should not be retryable for mutations")
	assert.False(t, isRetryable(fmt.Errorf("connection reset"), false), "connection reset should not be retryable for mutations")
	assert.False(t, isRetryable(fmt.Errorf("unexpected EOF"), false), "EOF should not be retryable for mutations")

	// API status code errors should still be retryable regardless of mode.
	assert.True(t, isRetryable(apiError(503, ""), false), "503 should be retryable even for mutations")
	assert.True(t, isRetryable(apiError(429, ""), false), "429 should be retryable even for mutations")
	assert.False(t, isRetryable(apiError(400, ""), false), "400 should not be retryable")
}

func TestRetryWithBackoff_PlainErrorNotRetried(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), fastOpts(), func(ctx context.Context) error {
		calls++
		return fmt.Errorf("not a transient error")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithRetry_NoRetryFlag(t *testing.T) {
	oldNoRetry := NoRetry
	defer func() { NoRetry = oldNoRetry }()

	NoRetry = true
	calls := 0
	err := WithRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return apiError(503, "")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls, "--no-retry should disable retries")
}

func TestWithRetry_MaxRetriesOverride(t *testing.T) {
	oldMaxRetries := MaxRetries
	oldNoRetry := NoRetry
	defer func() {
		MaxRetries = oldMaxRetries
		NoRetry = oldNoRetry
	}()

	NoRetry = false
	MaxRetries = 1
	calls := 0
	_ = WithRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return apiError(503, "")
	})
	assert.Equal(t, 2, calls, "MaxRetries=1 should allow 1 initial + 1 retry")
}

func TestWithIdempotentRetry_RetriesNetworkErrors(t *testing.T) {
	oldMaxRetries := MaxRetries
	oldNoRetry := NoRetry
	defer func() {
		MaxRetries = oldMaxRetries
		NoRetry = oldNoRetry
	}()

	NoRetry = false
	MaxRetries = 2
	calls := 0
	_ = WithIdempotentRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return fmt.Errorf("connection reset by peer")
	})
	assert.Equal(t, 3, calls, "WithIdempotentRetry should retry network errors (1 initial + 2 retries)")
}

func TestWithRetry_DoesNotRetryNetworkErrors(t *testing.T) {
	oldMaxRetries := MaxRetries
	oldNoRetry := NoRetry
	defer func() {
		MaxRetries = oldMaxRetries
		NoRetry = oldNoRetry
	}()

	NoRetry = false
	MaxRetries = 2
	calls := 0
	_ = WithRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return fmt.Errorf("connection reset by peer")
	})
	assert.Equal(t, 1, calls, "WithRetry should NOT retry network errors for mutation safety")
}

func TestWithIdempotentRetry_NoRetryFlag(t *testing.T) {
	oldNoRetry := NoRetry
	defer func() { NoRetry = oldNoRetry }()

	NoRetry = true
	calls := 0
	err := WithIdempotentRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return apiError(503, "")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls, "--no-retry should disable retries for WithIdempotentRetry too")
}

func TestAddJitter(t *testing.T) {
	base := 1 * time.Second
	for i := 0; i < 100; i++ {
		result := addJitter(base)
		require.GreaterOrEqual(t, result, base, "jitter should not reduce duration")
		require.LessOrEqual(t, result, base+base/4, "jitter should not exceed 25%%")
	}
}

func TestRetryAfterDelay(t *testing.T) {
	t.Run("no API error", func(t *testing.T) {
		d, ok := retryAfterDelay(fmt.Errorf("plain"))
		assert.Equal(t, time.Duration(0), d)
		assert.False(t, ok)
	})
	t.Run("non-429", func(t *testing.T) {
		d, ok := retryAfterDelay(apiError(503, "5"))
		assert.Equal(t, time.Duration(0), d)
		assert.False(t, ok)
	})
	t.Run("429 with seconds", func(t *testing.T) {
		d, ok := retryAfterDelay(apiError(429, "5"))
		assert.Equal(t, 5*time.Second, d)
		assert.True(t, ok)
	})
	t.Run("429 with zero seconds", func(t *testing.T) {
		d, ok := retryAfterDelay(apiError(429, "0"))
		assert.Equal(t, time.Duration(0), d)
		assert.True(t, ok, "Retry-After: 0 is valid and means retry immediately")
	})
	t.Run("429 without header", func(t *testing.T) {
		d, ok := retryAfterDelay(apiError(429, ""))
		assert.Equal(t, time.Duration(0), d)
		assert.False(t, ok)
	})
	t.Run("429 with HTTP-date in future", func(t *testing.T) {
		future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
		d, ok := retryAfterDelay(apiError(429, future))
		assert.True(t, ok)
		assert.Greater(t, d, 5*time.Second, "should parse future HTTP-date")
		assert.LessOrEqual(t, d, 11*time.Second)
	})
	t.Run("429 with HTTP-date in past", func(t *testing.T) {
		past := time.Now().Add(-10 * time.Second).UTC().Format(http.TimeFormat)
		d, ok := retryAfterDelay(apiError(429, past))
		assert.Equal(t, time.Duration(0), d)
		assert.True(t, ok, "past HTTP-date is valid, means retry immediately")
	})
	t.Run("429 with invalid value", func(t *testing.T) {
		d, ok := retryAfterDelay(apiError(429, "not-a-number"))
		assert.Equal(t, time.Duration(0), d)
		assert.False(t, ok)
	})
}

func TestRetryWithBackoff_RetryAfterCappedAtMaxDelay(t *testing.T) {
	calls := 0
	start := time.Now()
	opts := RetryOpts{
		MaxRetries:        1,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          500 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	err := RetryWithBackoff(context.Background(), opts, func(ctx context.Context) error {
		calls++
		if calls == 1 {
			return apiError(429, "60") // Server says wait 60s
		}
		return nil
	})
	elapsed := time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, 2, calls)
	// Should have waited ~500ms (MaxDelay cap), not 60s
	assert.Less(t, elapsed, 2*time.Second)
}

func TestLogRetry_QuietMode(t *testing.T) {
	output.SetVerbosity("quiet")
	defer output.SetVerbosity("normal")

	// Should not panic and should not produce output; exercises the quiet branch
	logRetry(1, 3, 100*time.Millisecond, fmt.Errorf("test error"))
	logRetry(3, 3, 100*time.Millisecond, fmt.Errorf("final retry error"))
}

func TestLogRetry_VerboseMode(t *testing.T) {
	output.SetVerbosity("verbose")
	defer output.SetVerbosity("normal")

	// Should not panic; just exercises the verbose branch
	logRetry(1, 3, 100*time.Millisecond, fmt.Errorf("test error"))
}

func TestLogRetry_NonVerboseEarlyAttempt(t *testing.T) {
	output.SetVerbosity("normal")

	// Early attempt, non-verbose: should not log (exercises the skip branch)
	logRetry(1, 3, 100*time.Millisecond, fmt.Errorf("test error"))
}

// fastOpts returns retry options with minimal delays for fast tests.
func fastOpts() RetryOpts {
	return RetryOpts{
		MaxRetries:        3,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
}
