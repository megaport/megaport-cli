package managed_account

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for managed account commands
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "managed-account"
}

// RegisterCommands adds managed account commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new managed account module
func NewModule() *Module {
	return &Module{}
}
