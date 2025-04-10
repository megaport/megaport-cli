package config

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for config
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "config"
}

// RegisterCommands adds the config command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new config module
func NewModule() *Module {
	return &Module{}
}
