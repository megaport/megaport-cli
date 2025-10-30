//go:build js && wasm
// +build js,wasm

package megaport

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/registry"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Common variables and declarations for WASM builds
var (
	noColor      bool
	outputFormat string

	// rootCmd is the root command for the CLI
	rootCmd = &cobra.Command{
		Use:           "megaport-cli",
		Short:         "A CLI tool to interact with the Megaport API",
		SilenceUsage:  true,  // Don't show usage on errors
		SilenceErrors: false, // Do show errors
		// Long will be set by the help builder later
		Run: func(cmd *cobra.Command, args []string) {
			// If we reach here with args, it means an unknown command was used
			if len(args) > 0 {
				// Special message for config commands since they're not available in WASM
				if args[0] == "config" {
					fmt.Fprintf(cmd.OutOrStderr(), "Error: the 'config' command is not available in the browser version\n\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Configuration is managed through the web UI login form.\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Use the --env flag to specify environment: megaport-cli ports list --env staging\n")
					return
				}
				// Show error for other unknown commands
				fmt.Fprintf(cmd.OutOrStderr(), "Error: unknown command %q for %q\n\n", args[0], cmd.CommandPath())
				fmt.Fprintf(cmd.OutOrStderr(), "Run 'megaport-cli --help' for usage\n")
				return
			}
			// No args - just show help
			cmd.Help()
		},
	}

	// moduleRegistry holds all command modules
	moduleRegistry *registry.Registry
)

// InitializeCommon performs initialization steps common to all platforms
func InitializeCommon() {
	// Initialize module registry
	moduleRegistry = registry.NewRegistry()

	// Register all modules (WASM version excludes config)
	registerModules()

	// Setup persistent flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", utils.FormatTable,
		"Output format (table, json, csv, xml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorful output")
	rootCmd.PersistentFlags().StringVar(&utils.Env, "env", "", "Environment to use (prod, dev, or staging)")
}
