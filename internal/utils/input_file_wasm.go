//go:build js && wasm

package utils

import "errors"

// errBrowserFileInput backs the --json-file path in the browser build, where
// there is no OS filesystem to read from. The host should read the file with
// the browser File API and pass its content inline. ReadJSONInput also backs a
// --resource-tags-file path (NAT gateway), so the hint names an example flag
// rather than a single flag that would be wrong for one of the callers.
var errBrowserFileInput = errors.New("file input is not supported in the browser; use the corresponding inline flag instead (e.g. --json)")

// errBrowserTagsFileInput backs the tags-file reader (readTagsFile), which is
// shared by --resource-tags-file, --tags-file, and the tags --json-file paths.
// Their inline counterparts differ (--resource-tags / --tags / --json), so the
// hint names an example rather than a single flag that would misdirect one path.
var errBrowserTagsFileInput = errors.New("file input is not supported in the browser; use the corresponding inline flag instead (e.g. --resource-tags)")

func readInputFile(_ string) ([]byte, error) { return nil, errBrowserFileInput }

func readTagsFile(_ string) ([]byte, error) { return nil, errBrowserTagsFileInput }
