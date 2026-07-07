//go:build !js || !wasm

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// readInputFile reads a --json-file path from the OS filesystem, applying the
// same path-traversal, regular-file, and size guards as readTagsFile.
func readInputFile(path string) ([]byte, error) {
	return readFileGuarded(path, "failed to read JSON file")
}

// readTagsFile reads a --resource-tags-file / --tags-file path from the OS
// filesystem under the same guards as readInputFile.
func readTagsFile(path string) ([]byte, error) {
	return readFileGuarded(path, "failed to read file")
}

// readFileGuarded reads a user-supplied file path safely: it cleans the path,
// rejects upward traversal, rejects non-regular files, and enforces a 1 MiB
// size limit while reading (open + stat + LimitReader avoids a TOCTOU race).
// readErrPrefix labels I/O failures so callers keep their existing wording.
func readFileGuarded(path, readErrPrefix string) ([]byte, error) {
	clean := filepath.Clean(path)
	// Reject paths that still navigate above the current directory after cleaning.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid file path %q: path traversal not allowed", path)
	}
	f, err := os.Open(clean)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", readErrPrefix, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", readErrPrefix, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("file %q is not a regular file", path)
	}
	if info.Size() > maxInputFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, info.Size())
	}
	data, err := io.ReadAll(io.LimitReader(f, maxInputFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", readErrPrefix, err)
	}
	if len(data) > maxInputFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, len(data))
	}
	return data, nil
}
