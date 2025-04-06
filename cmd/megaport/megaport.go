package megaport

import (
	"fmt"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/commands/completion"
	"github.com/megaport/megaport-cli/internal/commands/generate_docs"
	"github.com/megaport/megaport-cli/internal/commands/locations"
	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/commands/megaport"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/commands/partners"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/commands/servicekeys"
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// getIsColorDisabled checks whether color output should be disabled
func getIsColorDisabled() bool {
	// Check if NO_COLOR environment variable is set (standard for disabling color)
	_, noColorEnv := os.LookupEnv("NO_COLOR")

	return noColor || noColorEnv
}

func init() {
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
		// This function runs after flags are parsed, so noColor will have the correct value
		isColorDisabled := getIsColorDisabled()

		// Create a copy of the original Long description
		originalLong := cmd.Long

		// For the root command, regenerate the help text completely
		if cmd == rootCmd {
			rootHelp := getRootHelpBuilder(isColorDisabled)
			cmd.Long = rootHelp.Build(rootCmd)
		} else if cmd.Long != "" {
			// For non-root commands, modify the existing help text only if there is a Long description
			helpBuilder := &help.CommandHelpBuilder{
				CommandName:  cmd.UseLine(),
				ShortDesc:    cmd.Short,
				LongDesc:     originalLong,
				DisableColor: isColorDisabled,
			}
			cmd.Long = helpBuilder.Build(rootCmd)
		}

		// Call the original help function directly
		originalHelpFunc(cmd, args)

		// Restore the original Long description
		cmd.Long = originalLong
	})

	// Set the initial root command help text
	rootCmd.Long = getRootHelpBuilder(false).Build(rootCmd)

	// Add subcommands
	generate_docs.AddCommandsTo(rootCmd)
	completion.AddCommandsTo(rootCmd)
	locations.AddCommandsTo(rootCmd)
	megaport.AddCommandsTo(rootCmd)
	ports.AddCommandsTo(rootCmd)
	vxc.AddCommandsTo(rootCmd)
	mcr.AddCommandsTo(rootCmd)
	mve.AddCommandsTo(rootCmd)
	partners.AddCommandsTo(rootCmd)
	servicekeys.AddCommandsTo(rootCmd)
}
