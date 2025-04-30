package partners

import (
	"github.com/spf13/cobra"
)

type Module struct{}

func (m *Module) Name() string {
	return "partners"
}

func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

func NewModule() *Module {
	return &Module{}
}
