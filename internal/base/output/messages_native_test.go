//go:build !wasm

package output

import (
	"encoding/json"
	"testing"

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

func TestPrintErrorSuppressedInJSON(t *testing.T) {
	saveOutputFormat(t)
	SetOutputFormat("json")
	ResetErrorPrinted()

	stdout := captureStdout(t, func() {
		stderr := captureStderr(t, func() {
			PrintError("boom %d", true, 1)
		})
		// Human line is suppressed in JSON mode; the wrapper emits the
		// structured envelope instead.
		assert.Empty(t, stderr)
	})
	assert.Empty(t, stdout)
	// The flag is still set so the wrapper knows an error was reported.
	assert.True(t, ErrorWasPrinted())
}

func TestPrintErrorMarksAndRoutesToStderrForTable(t *testing.T) {
	saveOutputFormat(t)
	SetOutputFormat("table")
	ResetErrorPrinted()
	assert.False(t, ErrorWasPrinted())

	stdout := captureStdout(t, func() {
		stderr := captureStderr(t, func() {
			PrintError("boom", true)
		})
		assert.Contains(t, stderr, "✗ boom")
	})
	assert.Empty(t, stdout, "PrintError must not pollute stdout")
	assert.True(t, ErrorWasPrinted())
}

func TestPrintErrorPlainAlwaysWritesStderr(t *testing.T) {
	saveOutputFormat(t)
	// Even with the global format set to json, PrintErrorPlain writes
	// unconditionally: it must not consult the output format.
	SetOutputFormat("json")
	ResetErrorPrinted()

	stderr := captureStderr(t, func() {
		PrintErrorPlain("invalid output format: JSON", true)
	})
	assert.Equal(t, "✗ invalid output format: JSON\n", stderr)
	// PrintErrorPlain does not mark the printed-error flag.
	assert.False(t, ErrorWasPrinted())
}

func TestResetErrorPrinted(t *testing.T) {
	saveOutputFormat(t)
	SetOutputFormat("table")
	markErrorPrinted()
	assert.True(t, ErrorWasPrinted())
	ResetErrorPrinted()
	assert.False(t, ErrorWasPrinted())
}
