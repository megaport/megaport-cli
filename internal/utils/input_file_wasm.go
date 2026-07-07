//go:build js && wasm

package utils

import "errors"

// errBrowserFileInput backs every file-path input in the browser build, where
// there is no OS filesystem to read from: readInputFile's --json-file path
// (also used by --resource-tags-file via ReadJSONInput in NAT gateway) and
// readTagsFile's --resource-tags-file / --json-file paths. Their inline
// counterparts differ (--json vs --resource-tags), so the message names no
// specific flag.
var errBrowserFileInput = errors.New("file input is not supported in the browser; use the corresponding inline flag instead")

func readInputFile(_ string) ([]byte, error) { return nil, errBrowserFileInput }

func readTagsFile(_ string) ([]byte, error) { return nil, errBrowserFileInput }
