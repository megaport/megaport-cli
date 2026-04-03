//go:build js && wasm
// +build js,wasm

package utils

import (
	"context"
	"fmt"
)

// WatchLoop is not supported in the browser.
func WatchLoop(_ context.Context, _ WatchConfig, _ func(ctx context.Context) (string, error)) error {
	return fmt.Errorf("watch mode is not supported in the browser")
}
