package status

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for the status command.
type Module struct{}

// Name returns the module name.
func (m *Module) Name() string {
	return "status"
}

// RegisterCommands adds the status command to the root command.
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new status module.
func NewModule() *Module {
	return &Module{}
}
