package mcr

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for mcr
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "mcr"
}

// RegisterCommands adds the mcr command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new mcr module
func NewModule() *Module {
	return &Module{}
}
