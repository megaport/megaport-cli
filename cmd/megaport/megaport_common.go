// This file has no build tags so it's included in all builds

package megaport

import (
	"github.com/megaport/megaport-cli/internal/base/registry"
	"github.com/megaport/megaport-cli/internal/commands/completion"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/commands/generate_docs"
	"github.com/megaport/megaport-cli/internal/commands/locations"
	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/commands/partners"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/commands/servicekeys"
	"github.com/megaport/megaport-cli/internal/commands/version"
	"github.com/megaport/megaport-cli/internal/commands/vxc"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Common variables and declarations needed by both WASM and non-WASM builds
var (
	noColor      bool
	outputFormat string

	// rootCmd is the root command for the CLI
	rootCmd = &cobra.Command{
		Use:   "megaport-cli",
		Short: "A CLI tool to interact with the Megaport API",
		// Long will be set by the help builder later
	}

	// moduleRegistry holds all command modules
	moduleRegistry *registry.Registry
)

// registerModules registers all command modules with the registry
func registerModules() {
	// Register all modules
	moduleRegistry.Register(version.NewModule())
	moduleRegistry.Register(ports.NewModule())
	moduleRegistry.Register(vxc.NewModule())
	moduleRegistry.Register(mcr.NewModule())
	moduleRegistry.Register(mve.NewModule())
	moduleRegistry.Register(locations.NewModule())
	moduleRegistry.Register(partners.NewModule())
	moduleRegistry.Register(servicekeys.NewModule())
	moduleRegistry.Register(generate_docs.NewModule())
	moduleRegistry.Register(completion.NewModule())
	moduleRegistry.Register(config.NewModule())
}

// InitializeCommon performs initialization steps common to all platforms
func InitializeCommon() {
	// Initialize module registry
	moduleRegistry = registry.NewRegistry()

	// Register all modules
	registerModules()

	// Setup persistent flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", utils.FormatTable,
		"Output format (table, json, csv, xml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorful output")
	rootCmd.PersistentFlags().StringVar(&utils.Env, "env", "", "Environment to use (prod, dev, or staging)")
}
