package megaport

import (
	"fmt"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/base/registry"
	"github.com/megaport/megaport-cli/internal/commands/completion"
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

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "megaport-cli",
	Short: "A CLI tool to interact with the Megaport API",
	// Long will be set by the help builder later
}

var outputFormat string

// moduleRegistry holds all command modules
var moduleRegistry *registry.Registry

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Enable pretty tables by default but disable for tests
	output.UsePrettyTables = true

	// Detect test environment
	for _, arg := range os.Args {
		if strings.Contains(arg, "test.run") {
			output.UsePrettyTables = false
			break
		}
	}
	// Initialize module registry
	moduleRegistry = registry.NewRegistry()

	// Register all modules
	registerModules()

	// Setup persistent flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", utils.FormatTable,
		fmt.Sprintf("Output format (%s)", strings.Join(utils.ValidFormats, ", ")))
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorful output")
	rootCmd.PersistentFlags().StringVar(&utils.Env, "env", "", "Environment to use (prod, dev, or staging)")

	// Set up validation for the output format
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		format := strings.ToLower(outputFormat)
		for _, validFormat := range utils.ValidFormats {
			if format == validFormat {
				outputFormat = format
				return nil
			}
		}
		return fmt.Errorf("invalid output format: %s. Must be one of: %s",
			outputFormat, strings.Join(utils.ValidFormats, ", "))
	}

	// Store the original help function so we can call it when needed
	originalHelpFunc := rootCmd.HelpFunc()

	// Define the help text configuration once to avoid duplication
	getRootHelpBuilder := func(disableColor bool) *help.CommandHelpBuilder {
		return &help.CommandHelpBuilder{
			CommandName: "megaport-cli",
			ShortDesc:   "A CLI tool to interact with the Megaport API",
			LongDesc:    "Megaport CLI provides a command line interface to interact with the Megaport API.\n\nThe CLI allows you to manage Megaport resources such as ports, VXCs, MCRs, MVEs, service keys, and more.",
			OptionalFlags: map[string]string{
				"--no-color": "Disable colored output",
				"--output":   "Output format (json, yaml, table, csv, xml)",
				"--help":     "Show help for any command",
				"--env":      "Environment to use (production, staging, development)",
			},
			Examples: []string{
				"megaport-cli ports list",
				"megaport-cli vxc buy --interactive",
				"megaport-cli mcr get [mcrUID]",
				"megaport-cli locations list",
			},
			ImportantNotes: []string{
				"Use the --help flag with any command to see specific usage information",
				"Authentication is handled via the MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY environment variables",
				"By default, the CLI connects to the Megaport production environment",
				"Set the MEGAPORT_ENDPOINT environment variable to connect to a different environment",
			},
			DisableColor: disableColor,
		}
	}

	// Create a help function that runs the help.CommandHelpBuilder with the current noColor setting
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		isColorDisabled := false

		// Check raw command line args directly
		for _, arg := range os.Args {
			if arg == "--no-color" || arg == "-no-color" {
				isColorDisabled = true
				break
			}
		}

		// Also check environment
		if _, exists := os.LookupEnv("NO_COLOR"); exists {
			isColorDisabled = true
		}

		if isColorDisabled {
			output.UsePrettyTables = false
		}

		// For the root command, regenerate the help text completely
		if cmd == rootCmd {
			rootHelp := getRootHelpBuilder(isColorDisabled)
			cmd.Long = rootHelp.Build(rootCmd)
		} else if cmd.Long != "" {
			// For non-root commands, modify the existing help text only if there is a Long description
			helpBuilder := &help.CommandHelpBuilder{
				CommandName:  cmd.UseLine(),
				ShortDesc:    cmd.Short,
				LongDesc:     cmd.Long,
				DisableColor: isColorDisabled,
			}
			cmd.Long = helpBuilder.Build(rootCmd)
		}

		// Call the original help function
		originalHelpFunc(cmd, args)
	})

	// Set the initial root command help text
	rootCmd.Long = getRootHelpBuilder(false).Build(rootCmd)

	// Register all commands from the modules
	moduleRegistry.RegisterAll(rootCmd)
}

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
}
