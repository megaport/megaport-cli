//go:build js && wasm
// +build js,wasm

package megaport

import (
	"fmt"
	"io"

	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/spf13/cobra"
)

func ExecuteWithArgs(args []string) {
	// Debug
	fmt.Printf("WASM ExecuteWithArgs: args=%v\n", args)

	// Direct output to our WASM buffer
	rootCmd.SetOut(wasm.WasmOutputBuffer)
	rootCmd.SetErr(wasm.WasmOutputBuffer)

	// Enable traversal for subcommand flags
	rootCmd.PersistentFlags().ParseErrorsWhitelist.UnknownFlags = true
	rootCmd.TraverseChildren = true

	// Enable subcommand traversal for ALL commands, not just root
	enableTraversalForAllCommands(rootCmd)

	// Properly handle args - if first arg is the program name, skip it
	argsToUse := args
	if len(args) > 0 && (args[0] == "megaport-cli" || args[0] == "./megaport-cli") {
		argsToUse = args[1:]
	}

	// Debug the actual args we're using
	fmt.Printf("WASM using args for command: %v\n", argsToUse)

	// Set the args on the root command
	rootCmd.SetArgs(argsToUse)

	// Execute and capture errors
	err := rootCmd.Execute()

	// Debug the command result - only report errors
	// The command output is already in WasmOutputBuffer
	if err != nil {
		fmt.Fprintf(wasm.WasmOutputBuffer, "Error executing command: %v\n", err)
	}
}

// New helper function to enable traversal on all commands
func enableTraversalForAllCommands(cmd *cobra.Command) {
	cmd.TraverseChildren = true
	cmd.Flags().ParseErrorsWhitelist.UnknownFlags = true

	for _, subCmd := range cmd.Commands() {
		enableTraversalForAllCommands(subCmd)
	}
}

func EnsureRootCommandOutput(writer io.Writer) {
	rootCmd.SetOut(writer)
	rootCmd.SetErr(writer)
}

func init() {
	// Initialize common components
	InitializeCommon()

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
		// For the root command, regenerate the help text completely
		if cmd == rootCmd {
			rootHelp := getRootHelpBuilder(noColor)
			cmd.Long = rootHelp.Build(rootCmd)
		} else if cmd.Long != "" {
			// For non-root commands, modify the existing help text only if there is a Long description
			helpBuilder := &help.CommandHelpBuilder{
				CommandName:  cmd.UseLine(),
				ShortDesc:    cmd.Short,
				LongDesc:     cmd.Long,
				DisableColor: noColor,
			}
			cmd.Long = helpBuilder.Build(rootCmd)
		}

		// Call the original help function
		originalHelpFunc(cmd, args)
	})

	// Set the initial root command help text
	rootCmd.Long = getRootHelpBuilder(false).Build(rootCmd)

	moduleRegistry.RegisterAll(rootCmd)
}
