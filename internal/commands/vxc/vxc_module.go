package vxc

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for vxc
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "vxc"
}

// RegisterCommands adds the vxc command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new vxc module
func NewModule() *Module {
	return &Module{}
}
