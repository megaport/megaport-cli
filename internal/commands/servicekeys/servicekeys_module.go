package servicekeys

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for servicekeys
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "servicekeys"
}

// RegisterCommands adds the servicekeys command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new servicekeys module
func NewModule() *Module {
	return &Module{}
}
