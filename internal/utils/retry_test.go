package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

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
		{"connection reset", fmt.Errorf("read tcp: connection reset by peer"), true},
		{"connection refused", fmt.Errorf("dial tcp: connection refused"), true},
		{"EOF", fmt.Errorf("unexpected EOF"), true},
		{"i/o timeout", fmt.Errorf("i/o timeout"), true},
		{"plain error", fmt.Errorf("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retryable, isRetryable(tt.err))
		})
	}
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
		assert.Equal(t, time.Duration(0), retryAfterDelay(fmt.Errorf("plain")))
	})
	t.Run("non-429", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), retryAfterDelay(apiError(503, "5")))
	})
	t.Run("429 with seconds", func(t *testing.T) {
		assert.Equal(t, 5*time.Second, retryAfterDelay(apiError(429, "5")))
	})
	t.Run("429 without header", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), retryAfterDelay(apiError(429, "")))
	})
	t.Run("429 with HTTP-date in future", func(t *testing.T) {
		future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
		d := retryAfterDelay(apiError(429, future))
		assert.Greater(t, d, 5*time.Second, "should parse future HTTP-date")
		assert.LessOrEqual(t, d, 11*time.Second)
	})
	t.Run("429 with HTTP-date in past", func(t *testing.T) {
		past := time.Now().Add(-10 * time.Second).UTC().Format(http.TimeFormat)
		assert.Equal(t, time.Duration(0), retryAfterDelay(apiError(429, past)))
	})
	t.Run("429 with invalid value", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), retryAfterDelay(apiError(429, "not-a-number")))
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

func TestLogRetry_VerboseMode(t *testing.T) {
	oldVerbose := Verbose
	defer func() { Verbose = oldVerbose }()

	Verbose = true
	// Should not panic; just exercises the verbose branch
	logRetry(1, 3, 100*time.Millisecond, fmt.Errorf("test error"))
}

func TestLogRetry_NonVerboseEarlyAttempt(t *testing.T) {
	oldVerbose := Verbose
	defer func() { Verbose = oldVerbose }()

	Verbose = false
	// Early attempt, non-verbose: should not log (exercises the skip branch)
	logRetry(1, 3, 100*time.Millisecond, fmt.Errorf("test error"))
}

func TestIsVerboseMode(t *testing.T) {
	oldVerbose := Verbose
	defer func() { Verbose = oldVerbose }()

	Verbose = false
	assert.False(t, isVerboseMode())
	Verbose = true
	assert.True(t, isVerboseMode())
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
