//go:build js && wasm
// +build js,wasm

package megaport

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/base/registry"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Common variables and declarations for WASM builds
var (
	noColor      bool
	noHeader     bool
	outputFormat string
	quiet        bool
	verbose      bool

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
				// Special messages for commands not available in WASM
				switch args[0] {
				case "config":
					fmt.Fprintf(cmd.OutOrStderr(), "Error: the 'config' command is not available in the browser version\n\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Configuration is managed through the web UI login form.\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Use the --env flag to specify environment: megaport-cli ports list --env staging\n")
					return
				case "completion":
					fmt.Fprintf(cmd.OutOrStderr(), "Error: the 'completion' command is not available in the browser version\n\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Shell completion is not applicable in a browser environment.\n")
					return
				case "generate-docs":
					fmt.Fprintf(cmd.OutOrStderr(), "Error: the 'generate-docs' command is not available in the browser version\n\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Documentation generation is a development tool, not available in WASM.\n")
					return
				case "version":
					fmt.Fprintf(cmd.OutOrStderr(), "Error: the 'version' command is not available in the browser version\n\n")
					fmt.Fprintf(cmd.OutOrStderr(), "Version information is managed by the web application.\n")
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
		"Output format (table, json, csv, xml, go-template; go-template not supported in browser version)")
	rootCmd.PersistentFlags().String("template", "", "Go template string for --output go-template (not supported in browser version)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorful output")
	rootCmd.PersistentFlags().StringVar(&utils.Env, "env", "", "Environment to use (prod, dev, or staging)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational output, only show errors and data")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show additional debug information")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Request timeout duration (e.g., 30s, 2m, 5m); 0 uses the internal default of 90s")
	rootCmd.PersistentFlags().String("fields", "", "Comma-separated list of fields to include in output (e.g., uid,name,status); use an unknown name to list available fields")
	rootCmd.PersistentFlags().String("query", "", "JMESPath query to filter JSON output (requires --output json)")
	rootCmd.PersistentFlags().BoolVar(&utils.NoRetry, "no-retry", false, "Disable automatic retry on transient API failures")
	rootCmd.PersistentFlags().IntVar(&utils.MaxRetries, "max-retries", 3, "Maximum number of retries for transient API failures")
	rootCmd.PersistentFlags().BoolVar(&noHeader, "no-header", false, "Suppress table and CSV column headers (useful for scripting)")
	rootCmd.MarkFlagsMutuallyExclusive("quiet", "verbose")

	// Validate retry flags in WASM builds too.
	existingPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		output.SetNoHeader(noHeader)
		if utils.MaxRetries < 0 {
			return fmt.Errorf("--max-retries must be >= 0, got %d", utils.MaxRetries)
		}
		if existingPreRunE != nil {
			return existingPreRunE(cmd, args)
		}
		return nil
	}
}
