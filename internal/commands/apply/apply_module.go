package apply

import "github.com/spf13/cobra"

// Module implements registry.Module for the apply command.
type Module struct{}

func (m *Module) Name() string { return "apply" }

func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

func NewModule() *Module { return &Module{} }
