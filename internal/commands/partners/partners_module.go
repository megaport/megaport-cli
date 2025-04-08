package partners

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for partners
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "partners"
}

// RegisterCommands adds the partners command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new partners module
func NewModule() *Module {
	return &Module{}
}
