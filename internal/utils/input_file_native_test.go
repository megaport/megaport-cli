//go:build !js || !wasm

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInputFile_Native_ReadsContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"name":"port"}`), 0o600))

	data, err := readInputFile(path)
	require.NoError(t, err)
	assert.Equal(t, `{"name":"port"}`, string(data))
}

func TestReadInputFile_Native_MissingFile(t *testing.T) {
	_, err := readInputFile(filepath.Join(t.TempDir(), "does-not-exist.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read JSON file")
}

func TestReadJSONInput_Native_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "order.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"term":12}`), 0o600))

	data, err := ReadJSONInput("", path)
	require.NoError(t, err)
	assert.Equal(t, `{"term":12}`, string(data))
}
