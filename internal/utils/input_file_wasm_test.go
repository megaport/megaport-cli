//go:build js && wasm

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wasmJSONFileMsg = "file input is not supported in the browser; use the corresponding inline flag instead"
	wasmTagsFileMsg = "file input is not supported in the browser; use the corresponding inline flag instead"
)

func TestReadInputFile_WASM_Unsupported(t *testing.T) {
	_, err := readInputFile("/some/path.json")
	require.Error(t, err)
	assert.Equal(t, wasmJSONFileMsg, err.Error())
}

func TestReadTagsFile_WASM_Unsupported(t *testing.T) {
	_, err := readTagsFile("/some/tags.json")
	require.Error(t, err)
	assert.Equal(t, wasmTagsFileMsg, err.Error())
}

func TestReadJSONInput_WASM_FileFails(t *testing.T) {
	_, err := ReadJSONInput("", "/some/order.json")
	require.Error(t, err)
	assert.Equal(t, wasmJSONFileMsg, err.Error())
}

func TestReadJSONInput_WASM_InlineStillWorks(t *testing.T) {
	data, err := ReadJSONInput(`{"term":12}`, "/ignored/path.json")
	require.NoError(t, err)
	assert.Equal(t, `{"term":12}`, string(data))
}
