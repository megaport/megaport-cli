//go:build !js && !wasm
// +build !js,!wasm

package megaport

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/spf13/cobra"
)

func init() {
	// Initialize common components
	InitializeCommon()

	// Apply non-WASM specific initialization
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		applyDefaultSettings(cmd)
		format := strings.ToLower(outputFormat)
		for _, validFormat := range utils.ValidFormats {
			if format == validFormat {
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
				"Set the MEGAPORT_ENVIRONMENT environment variable to connect to a different environment",
			},
			DisableColor: disableColor,
		}
	}

	// Create a help function that runs the help.CommandHelpBuilder with the current noColor setting
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		var isColorDisabled bool

		if noColor {
			isColorDisabled = true
		}

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

// Apply defaults from config to command flags
func applyDefaultSettings(cmd *cobra.Command) {
	manager, err := config.NewConfigManager()
	if err != nil {
		return // Silently continue if config can't be loaded
	}

	// Apply "no-color" default
	if !cmd.Flags().Changed("no-color") {
		if val, exists := manager.GetDefault("no-color"); exists {
			if boolVal, ok := val.(bool); ok {
				err := cmd.Flags().Set("no-color", fmt.Sprintf("%t", boolVal))
				if err != nil {
					log.Printf("Error setting no-color flag: %v\n", err)
					return
				}
				noColor = boolVal
			}
		}

	}

	// Apply "output" default
	if !cmd.Flags().Changed("output") {
		if val, exists := manager.GetDefault("output"); exists {
			if strVal, ok := val.(string); ok {
				err := cmd.Flags().Set("output", strVal)
				if err != nil {
					log.Printf("Error setting output flag: %v\n", err)
					return
				}
			}
		}
	}
}

// ExecuteWithArgs adds all child commands to the root command and executes with given args.
func ExecuteWithArgs(args []string) {
	// Save original args
	originalArgs := os.Args

	// Debug
	fmt.Printf("WASM ExecuteWithArgs: args=%v\n", args)

	// Set new args for this execution
	os.Args = args

	// Direct output to our WASM buffer
	rootCmd.SetOut(wasm.WasmOutputBuffer)
	rootCmd.SetErr(wasm.WasmOutputBuffer)

	// IMPORTANT: When executing in WASM, set this to ensure args are processed properly
	rootCmd.SetArgs(args[1:]) // Skip program name (args[0])

	// Execute and capture errors
	err := rootCmd.Execute()

	// Debug the command result
	if err != nil {
		fmt.Fprintf(wasm.WasmOutputBuffer, "Error executing command: %v\n", err)
	}

	// Restore original args
	os.Args = originalArgs
}

func EnsureRootCommandOutput(writer io.Writer) {
	rootCmd.SetOut(writer)
	rootCmd.SetErr(writer)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
