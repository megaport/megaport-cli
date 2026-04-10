package auth

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for authentication commands
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "auth"
}

// RegisterCommands adds authentication commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new auth module
func NewModule() *Module {
	return &Module{}
}
