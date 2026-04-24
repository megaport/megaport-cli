package nat_gateway

import (
	"github.com/spf13/cobra"
)

// Module implements the registry.Module interface for nat-gateway commands.
type Module struct{}

func (m *Module) Name() string {
	return "nat-gateway"
}

func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

func NewModule() *Module {
	return &Module{}
}
