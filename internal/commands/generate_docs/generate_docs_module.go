package generate_docs

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for generate-docs
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "generate_docs"
}

// RegisterCommands adds the generate-docs command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new generate-docs module
func NewModule() *Module {
	return &Module{}
}
