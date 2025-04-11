package utils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatCSV   = "csv"
	FormatXML   = "xml"
)

var (
	Env          string
	OutputFormat string
	NoColor      bool
	ValidFormats = []string{FormatTable, FormatJSON, FormatCSV, FormatXML}
)

func ShouldDisableColors() bool {
	// Check if NO_COLOR environment variable is set (standard for disabling color)
	_, noColorEnv := os.LookupEnv("NO_COLOR")

	// Check raw command line args - more reliable with --help
	for _, arg := range os.Args {
		if arg == "--no-color" || arg == "-no-color" {
			return true
		}
	}

	// Finally check the global variable
	return NoColor || noColorEnv
}

func GetCurrentEnv() string {
	if Env == "" {
		return "production" // Default to production if not specified
	}
	return Env
}

// WrapRunE wraps a RunE function to set SilenceUsage to true if an error occurs and formats the error message.
func WrapRunE(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			return fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag", cmd.Name(), err, cmd.Name(), args)
		}
		return nil
	}
}

// WrapColorAwareRunE combines WrapRunE and WrapCommandFunc to handle both error formatting
// and passing the noColor flag to command functions.
func WrapColorAwareRunE(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get noColor value from root command
		noColor, err := cmd.Root().PersistentFlags().GetBool("no-color")
		if err != nil {
			noColor = false // Default to color if flag not found
		}

		// Call the function with the noColor parameter
		err = fn(cmd, args, noColor)

		// Error handling from WrapRunE
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			return fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag",
				cmd.Name(), err, cmd.Name(), args)
		}
		return nil
	}
}

// WrapOutputFormatRunE handles both output format and color settings in command functions.
// This wrapper takes a function that needs both output format and noColor parameters.
func WrapOutputFormatRunE(fn func(cmd *cobra.Command, args []string, noColor bool, format string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get noColor value from root command
		noColor, err := cmd.Root().PersistentFlags().GetBool("no-color")
		if err != nil {
			noColor = false // Default to color if flag not found
		}

		// Get output format value from command
		format, err := cmd.Flags().GetString("output")
		if err != nil {
			format = FormatTable // Default to table format if flag not found
		}

		// Validate format
		validFormat := false
		for _, f := range ValidFormats {
			if format == f {
				validFormat = true
				break
			}
		}
		if !validFormat {
			// If an invalid format is provided, return an error
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return fmt.Errorf("invalid output format: %s. Must be one of: %v", format, ValidFormats)
		}

		// Call the function with both parameters
		err = fn(cmd, args, noColor, format)

		// Error handling from WrapRunE
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			return fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag",
				cmd.Name(), err, cmd.Name(), args)
		}
		return nil
	}
}
