package completion

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for completion
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "completion"
}

// RegisterCommands adds the completion command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new completion module
func NewModule() *Module {
	return &Module{}
}
