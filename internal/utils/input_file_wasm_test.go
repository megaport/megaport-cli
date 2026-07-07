//go:build js && wasm

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadInputFile_WASM_Unsupported(t *testing.T) {
	_, err := readInputFile("/some/path.json")
	require.Error(t, err)
	require.ErrorIs(t, err, errBrowserFileInput)
}

func TestReadTagsFile_WASM_Unsupported(t *testing.T) {
	_, err := readTagsFile("/some/tags.json")
	require.Error(t, err)
	require.ErrorIs(t, err, errBrowserFileInput)
}

func TestReadJSONInput_WASM_FileFails(t *testing.T) {
	_, err := ReadJSONInput("", "/some/order.json")
	require.Error(t, err)
	require.ErrorIs(t, err, errBrowserFileInput)
}

func TestReadJSONInput_WASM_InlineStillWorks(t *testing.T) {
	data, err := ReadJSONInput(`{"term":12}`, "/ignored/path.json")
	require.NoError(t, err)
	require.Equal(t, `{"term":12}`, string(data))
}
