//go:build !wasm
// +build !wasm

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
