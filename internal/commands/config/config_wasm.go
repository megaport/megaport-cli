//go:build js && wasm
// +build js,wasm

package config

import (
	"github.com/spf13/cobra"
)

// AddCommandsTo is a no-op for WASM builds since config commands are not supported
// Config profiles are not available in WASM - use session-based authentication via the browser UI instead
func AddCommandsTo(rootCmd *cobra.Command) {
	// Config commands are intentionally not registered in WASM builds
	// The WASM version uses session-based authentication managed by the browser/server
	// Users should use the login form in the browser UI for authentication
}
