package mve

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for mve
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "mve"
}

// RegisterCommands adds the mve command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new mve module
func NewModule() *Module {
	return &Module{}
}
