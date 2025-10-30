//go:build js && wasm
// +build js,wasm

package config

import (
	"github.com/spf13/cobra"
)

// Module is a no-op module for WASM builds
// Config commands are not supported in WASM - use session-based auth via browser UI
type Module struct{}

func (m *Module) Name() string {
	return "config"
}

func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	// No-op: Config commands are not registered in WASM builds
	// WASM uses session-based authentication via the browser UI
}

func NewModule() *Module {
	return &Module{}
}
