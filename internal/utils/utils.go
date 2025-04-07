package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
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
		return "prod" // Default to production if not specified
	}
	return Env
}

var Prompt = func(msg string, noColor bool) (string, error) {
	if !noColor {
		fmt.Print(color.BlueString(msg))
	} else {
		fmt.Print(msg)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

var ConfirmPrompt = func(question string, noColor bool) bool {
	var response string

	if !noColor {
		fmt.Print(color.YellowString("%s [y/N]: ", question))
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return false // Or handle the error as appropriate for your use case
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
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
