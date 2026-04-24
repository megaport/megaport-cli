package users

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for user management commands
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "users"
}

// RegisterCommands adds user management commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new users module
func NewModule() *Module {
	return &Module{}
}
