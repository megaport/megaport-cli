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

func TestReadJSONInput_Native_FileNotFound(t *testing.T) {
	_, err := ReadJSONInput("", filepath.Join(t.TempDir(), "does-not-exist.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read JSON file")
}

func TestReadTagsFile_Native_ReadsContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"env":"prod"}`), 0o600))

	data, err := readTagsFile(path)
	require.NoError(t, err)
	assert.Equal(t, `{"env":"prod"}`, string(data))
}

func TestReadTagsFile_Native_RejectsTraversal(t *testing.T) {
	_, err := readTagsFile("../secrets.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal not allowed")
}

func TestReadTagsFile_Native_RejectsNonRegularFile(t *testing.T) {
	_, err := readTagsFile(t.TempDir()) // a directory is not a regular file
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a regular file")
}

func TestReadTagsFile_Native_RejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.json")
	require.NoError(t, os.WriteFile(path, make([]byte, maxTagsFileSize+1), 0o600))

	_, err := readTagsFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}
