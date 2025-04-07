package ports

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for ports
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "ports"
}

// RegisterCommands adds the ports command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new ports module
func NewModule() *Module {
	return &Module{}
}
