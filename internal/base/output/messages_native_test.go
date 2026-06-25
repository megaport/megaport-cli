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

// stdout must carry only formatted data: every status/error message goes to
// stderr regardless of output format.
func TestPrintError_RoutesToStderrNotStdout(t *testing.T) {
	for _, format := range []string{"table", "csv", "xml"} {
		t.Run(format, func(t *testing.T) {
			SetOutputFormat(format)
			t.Cleanup(func() { ResetState() })

			var stdout string
			stderr := captureStderr(t, func() {
				stdout = CaptureStdout(func() {
					PrintError("boom %d", true, 1)
				})
			})
			assert.Empty(t, stdout, "stdout must stay clean")
			assert.Contains(t, stderr, "boom 1")
		})
	}
}

func TestPrintSuccessWarningInfo_RouteToStderrNotStdout(t *testing.T) {
	SetOutputFormat("table")
	t.Cleanup(func() { ResetState() })

	var stdout string
	stderr := captureStderr(t, func() {
		stdout = CaptureStdout(func() {
			PrintSuccess("ok", true)
			PrintWarning("careful", true)
			PrintInfo("fyi", true)
		})
	})
	assert.Empty(t, stdout, "stdout must stay clean")
	assert.Contains(t, stderr, "ok")
	assert.Contains(t, stderr, "careful")
	assert.Contains(t, stderr, "fyi")
}

// In json mode the RunE wrapper owns error output via the structured envelope,
// so PrintError must not emit a duplicate human line.
func TestPrintError_NoOpInJSON(t *testing.T) {
	SetOutputFormat("json")
	t.Cleanup(func() { ResetState() })

	var stdout string
	stderr := captureStderr(t, func() {
		stdout = CaptureOutput(func() {
			PrintError("boom", true)
		})
	})
	assert.Empty(t, stdout)
	assert.Empty(t, stderr, "PrintError must be a no-op in json mode")
}

func TestPrintError_SetsErrorEmittedLatch(t *testing.T) {
	SetOutputFormat("table")
	t.Cleanup(func() { ResetState() })

	ResetErrorEmitted()
	assert.False(t, ErrorEmitted())

	_ = captureStderr(t, func() {
		PrintError("boom", true)
	})
	assert.True(t, ErrorEmitted(), "PrintError should record that an error was shown")
}

func TestPrintError_DoesNotSetLatchInJSON(t *testing.T) {
	SetOutputFormat("json")
	t.Cleanup(func() { ResetState() })

	ResetErrorEmitted()
	_ = captureStderr(t, func() {
		PrintError("boom", true)
	})
	assert.False(t, ErrorEmitted(), "json no-op must not set the latch")
}

func TestResetState_ClearsErrorEmittedLatch(t *testing.T) {
	SetOutputFormat("table")
	t.Cleanup(func() { ResetState() })

	_ = captureStderr(t, func() { PrintError("boom", true) })
	assert.True(t, ErrorEmitted())
	ResetState()
	assert.False(t, ErrorEmitted(), "ResetState must clear the latch")
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
