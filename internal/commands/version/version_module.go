package version

import (
	"github.com/spf13/cobra"
)

// Module implements the cmdbuilder.Module interface for version
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "version"
}

// RegisterCommands adds the version command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new version module
func NewModule() *Module {
	return &Module{}
}
