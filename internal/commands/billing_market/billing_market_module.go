package billing_market

import (
	"github.com/spf13/cobra"
)

// Module implements the registry.Module interface for billing market
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "billing-market"
}

// RegisterCommands adds the billing-market command to the root command
func (m *Module) RegisterCommands(rootCmd *cobra.Command) {
	AddCommandsTo(rootCmd)
}

// NewModule creates a new billing market module
func NewModule() *Module {
	return &Module{}
}
