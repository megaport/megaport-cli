//go:build !wasm
// +build !wasm

package megaport

import (
	"github.com/megaport/megaport-cli/internal/base/registry"
	"github.com/megaport/megaport-cli/internal/commands/apply"
	"github.com/megaport/megaport-cli/internal/commands/auth"
	"github.com/megaport/megaport-cli/internal/commands/billing_market"
	"github.com/megaport/megaport-cli/internal/commands/completion"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/commands/generate_docs"
	"github.com/megaport/megaport-cli/internal/commands/ix"
	"github.com/megaport/megaport-cli/internal/commands/locations"
	"github.com/megaport/megaport-cli/internal/commands/managed_account"
	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/commands/partners"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/commands/product"
	"github.com/megaport/megaport-cli/internal/commands/servicekeys"
	"github.com/megaport/megaport-cli/internal/commands/status"
	"github.com/megaport/megaport-cli/internal/commands/topology"
	"github.com/megaport/megaport-cli/internal/commands/users"
	"github.com/megaport/megaport-cli/internal/commands/version"
	"github.com/megaport/megaport-cli/internal/commands/vxc"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Common variables and declarations needed by both WASM and non-WASM builds
var (
	noColor      bool
	outputFormat string
	quiet        bool
	verbose      bool

	// rootCmd is the root command for the CLI
	rootCmd = &cobra.Command{
		Use:           "megaport-cli",
		Short:         "A CLI tool to interact with the Megaport API",
		SilenceErrors: false, // Allow Cobra to print returned errors so failures are always visible
		// Long will be set by the help builder later
	}

	// moduleRegistry holds all command modules
	moduleRegistry *registry.Registry
)

// registerModules registers all command modules with the registry
func registerModules() {
	// Register all modules
	moduleRegistry.Register(auth.NewModule())
	moduleRegistry.Register(version.NewModule())
	moduleRegistry.Register(ports.NewModule())
	moduleRegistry.Register(vxc.NewModule())
	moduleRegistry.Register(mcr.NewModule())
	moduleRegistry.Register(mve.NewModule())
	moduleRegistry.Register(locations.NewModule())
	moduleRegistry.Register(partners.NewModule())
	moduleRegistry.Register(product.NewModule())
	moduleRegistry.Register(servicekeys.NewModule())
	moduleRegistry.Register(generate_docs.NewModule())
	moduleRegistry.Register(completion.NewModule())
	moduleRegistry.Register(ix.NewModule())
	moduleRegistry.Register(managed_account.NewModule())
	moduleRegistry.Register(config.NewModule())
	moduleRegistry.Register(billing_market.NewModule())
	moduleRegistry.Register(users.NewModule())
	moduleRegistry.Register(status.NewModule())
	moduleRegistry.Register(topology.NewModule())
	moduleRegistry.Register(apply.NewModule())
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
	rootCmd.PersistentFlags().StringVar(&utils.ProfileOverride, "profile", "", "Use a specific config profile for this command")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational output, only show errors and data")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show additional debug information")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Request timeout duration (e.g., 30s, 2m, 5m); 0 uses the internal default of 90s")
	rootCmd.PersistentFlags().String("fields", "", "Comma-separated list of fields to include in output (e.g., uid,name,status); use an unknown name to list available fields")
	rootCmd.PersistentFlags().String("query", "", "JMESPath query to filter JSON output (requires --output json)")
	rootCmd.PersistentFlags().BoolVar(&utils.NoRetry, "no-retry", false, "Disable automatic retry on transient API failures")
	rootCmd.PersistentFlags().IntVar(&utils.MaxRetries, "max-retries", 3, "Maximum number of retries for transient API failures")
	rootCmd.MarkFlagsMutuallyExclusive("quiet", "verbose")
}
