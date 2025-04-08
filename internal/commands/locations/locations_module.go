package locations

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for locations
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "locations"
}

// RegisterCommands adds all location commands to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new locations module
func NewModule() *Module {
	return &Module{}
}
