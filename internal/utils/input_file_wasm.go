//go:build js && wasm

package utils

import "errors"

// errBrowserFileInput backs the --json-file path in the browser build, where
// there is no OS filesystem to read from. The host should read the file with
// the browser File API and pass its content inline via --json.
var errBrowserFileInput = errors.New("file input is not supported in the browser; pass the JSON inline with --json")

// errBrowserTagsFileInput backs the --resource-tags-file path. Its inline
// counterpart is --resource-tags, not --json, so the hint stays generic rather
// than naming the wrong flag.
var errBrowserTagsFileInput = errors.New("file input is not supported in the browser; pass the tags inline instead")

func readInputFile(_ string) ([]byte, error) { return nil, errBrowserFileInput }

func readTagsFile(_ string) ([]byte, error) { return nil, errBrowserTagsFileInput }
