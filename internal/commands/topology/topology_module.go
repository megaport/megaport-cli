package topology

import "github.com/spf13/cobra"

// Module implements the registry.Module interface for the topology command
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "topology"
}

// RegisterCommands adds topology commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new topology module
func NewModule() *Module {
	return &Module{}
}
