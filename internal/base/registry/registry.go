package registry

import (
	"github.com/spf13/cobra"
)

// Module is the interface that all command modules must implement
type Module interface {
	// Name returns the module's name
	Name() string

	// RegisterCommands registers the module's commands with the root command
	RegisterCommands(rootCmd *cobra.Command)
}

// Registry keeps track of all registered modules
type Registry struct {
	modules []Module
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make([]Module, 0),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module Module) {
	r.modules = append(r.modules, module)
}

// RegisterAll registers all modules with the root command
func (r *Registry) RegisterAll(rootCmd *cobra.Command) {
	for _, module := range r.modules {
		module.RegisterCommands(rootCmd)
	}
}
