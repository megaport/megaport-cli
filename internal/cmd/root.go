package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "megaport-cli",
	Short: "A CLI tool to interact with the Megaport API",
	// Long will be set by the help builder later
}

const (
	formatTable = "table"
	formatJSON  = "json"
	formatCSV   = "csv"
	formatXML   = "xml"
)

var (
	env          string
	outputFormat string
	noColor      bool
	validFormats = []string{formatTable, formatJSON, formatCSV, formatXML}
)

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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", formatTable,
		fmt.Sprintf("Output format (%s)", strings.Join(validFormats, ", ")))
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorful output")

	// Add the environment flag if not already defined
	if rootCmd.PersistentFlags().Lookup("env") == nil {
		rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "production", "Environment to use (production, staging, development)")
	}

	// Set up validation for the output format
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		format := strings.ToLower(outputFormat)
		for _, validFormat := range validFormats {
			if format == validFormat {
				outputFormat = format
				return nil
			}
		}
		return fmt.Errorf("invalid output format: %s. Must be one of: %s",
			outputFormat, strings.Join(validFormats, ", "))
	}

	// Store the original help function so we can call it when needed
	originalHelpFunc := rootCmd.HelpFunc()

	// Create a help function that runs the CommandHelpBuilder with the current noColor setting
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// This function runs after flags are parsed, so noColor will have the correct value
		isColorDisabled := getIsColorDisabled()

		// Create a copy of the original Long description
		originalLong := cmd.Long

		// For the root command, regenerate the help text completely
		if cmd == rootCmd {
			rootHelp := &CommandHelpBuilder{
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
				DisableColor: isColorDisabled,
			}
			cmd.Long = rootHelp.Build()
		} else if cmd.Long != "" {
			// For non-root commands, modify the existing help text only if there is a Long description
			helpBuilder := &CommandHelpBuilder{
				CommandName:  cmd.UseLine(),
				ShortDesc:    cmd.Short,
				LongDesc:     originalLong,
				DisableColor: isColorDisabled,
			}
			cmd.Long = helpBuilder.Build()
		}

		// Call the original help function directly instead of using the parent
		originalHelpFunc(cmd, args)

		// Restore the original Long description
		cmd.Long = originalLong
	})

	// Set the initial root command help text
	rootHelp := &CommandHelpBuilder{
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
	}
	rootCmd.Long = rootHelp.Build()
}
