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

func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

func NewModule() *Module {
	return &Module{}
}
