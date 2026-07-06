//go:build js && wasm

package utils

import "errors"

// errBrowserFileInput backs the --json-file path in the browser build, where
// there is no OS filesystem to read from. The host should read the file with
// the browser File API and pass its content inline. ReadJSONInput also backs a
// --resource-tags-file path (NAT gateway), whose inline counterpart is
// --resource-tags rather than --json, so the message names no specific flag.
var errBrowserFileInput = errors.New("file input is not supported in the browser; use the corresponding inline flag instead")

// errBrowserTagsFileInput backs the tags-file reader (readTagsFile), which is
// shared by --resource-tags-file and a --json-file path used for resource-tags
// commands. Their inline counterparts differ (--resource-tags vs --json), so
// the message names no specific flag.
var errBrowserTagsFileInput = errors.New("file input is not supported in the browser; use the corresponding inline flag instead")

func readInputFile(_ string) ([]byte, error) { return nil, errBrowserFileInput }

func readTagsFile(_ string) ([]byte, error) { return nil, errBrowserTagsFileInput }
