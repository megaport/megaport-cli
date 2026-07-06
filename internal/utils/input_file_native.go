//go:build !js || !wasm

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// readInputFile reads a --json-file path from the OS filesystem.
func readInputFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}
	return data, nil
}

// readTagsFile reads a user-supplied JSON file path safely: it cleans the path,
// rejects upward traversal, rejects non-regular files, and enforces a 1 MiB
// size limit while reading (open + stat + LimitReader avoids a TOCTOU race).
func readTagsFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)
	// Reject paths that still navigate above the current directory after cleaning.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid file path %q: path traversal not allowed", path)
	}
	f, err := os.Open(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("file %q is not a regular file", path)
	}
	if info.Size() > maxTagsFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, info.Size())
	}
	data, err := io.ReadAll(io.LimitReader(f, maxTagsFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if len(data) > maxTagsFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, len(data))
	}
	return data, nil
}
