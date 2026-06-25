//go:build !wasm

package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintErrorJSON_GeneralError(t *testing.T) {
	out := captureStderr(t, func() {
		PrintErrorJSON(1, "something went wrong")
	})
	var env struct {
		Error struct {
			Code    int    `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal([]byte(out), &env))
	assert.Equal(t, 1, env.Error.Code)
	assert.Equal(t, "general_error", env.Error.Type)
	assert.Equal(t, "something went wrong", env.Error.Message)
}

func TestPrintErrorJSON_APIError(t *testing.T) {
	out := captureStderr(t, func() {
		PrintErrorJSON(4, "resource not found")
	})
	var env struct {
		Error struct {
			Code int    `json:"code"`
			Type string `json:"type"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal([]byte(out), &env))
	assert.Equal(t, 4, env.Error.Code)
	assert.Equal(t, "api_error", env.Error.Type)
}

func TestPrintErrorJSON_UsageError(t *testing.T) {
	out := captureStderr(t, func() {
		PrintErrorJSON(2, "bad flag value")
	})
	var env struct {
		Error struct {
			Code int    `json:"code"`
			Type string `json:"type"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal([]byte(out), &env))
	assert.Equal(t, 2, env.Error.Code)
	assert.Equal(t, "usage_error", env.Error.Type)
}

// TestSpinnerNonTTYDoesNotAnimate guards ESD-1515: in a non-TTY sink the spinner
// must print its status message once with no per-frame animation and no bare
// carriage-return / clear-line escapes that would litter CI logs.
func TestSpinnerNonTTYDoesNotAnimate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing-sensitive spinner test")
	}
	origTerm := isTerminalCached.Load()
	t.Cleanup(func() { SetIsTerminal(origTerm) })
	SetIsTerminal(false)

	const msg = "Validating Port order..."

	t.Run("Stop", func(t *testing.T) {
		out := captureStderr(t, func() {
			s := NewSpinner(true)
			s.Start(msg)
			time.Sleep(500 * time.Millisecond)
			s.Stop()
		})
		assert.Equal(t, 1, strings.Count(out, msg), "status message should appear exactly once, got: %q", out)
		assert.NotContains(t, out, "\r", "non-TTY output must not contain carriage returns")
		assert.NotContains(t, out, "\x1b[K", "non-TTY output must not contain clear-line escapes")
	})

	t.Run("StopWithSuccess", func(t *testing.T) {
		out := captureStderr(t, func() {
			s := NewSpinner(true)
			s.Start(msg)
			time.Sleep(500 * time.Millisecond)
			s.StopWithSuccess("Port order valid")
		})
		assert.Equal(t, 1, strings.Count(out, msg), "status message should appear exactly once, got: %q", out)
		assert.Contains(t, out, "✓ Port order valid", "final success line should still be emitted")
		assert.NotContains(t, out, "\r", "non-TTY output must not contain carriage returns")
		assert.NotContains(t, out, "\x1b[K", "non-TTY output must not contain clear-line escapes")
	})
}
