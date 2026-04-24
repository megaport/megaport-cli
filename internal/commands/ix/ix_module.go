package ix

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for IX commands
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "ix"
}

// RegisterCommands adds IX commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new IX module
func NewModule() *Module {
	return &Module{}
}
