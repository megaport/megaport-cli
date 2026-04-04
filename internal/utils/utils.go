package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatCSV   = "csv"
	FormatXML   = "xml"
)

var (
	// Env is the target environment (prod, dev, staging). Set once via flag
	// binding before command execution; read during login. Not protected by
	// a mutex because cobra flag parsing and command execution are sequential
	// on the main goroutine.
	Env string

	// ProfileOverride is the config profile name. Same set-once semantics as Env.
	ProfileOverride string

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

	return noColorEnv
}

func GetCurrentEnv() string {
	if Env == "" {
		return "production" // Default to production if not specified
	}
	return Env
}

// applyQueryFilter reads the --query persistent flag and calls output.SetOutputQuery.
// Returns the query string so RunE wrappers can validate it against the selected output format.
func applyQueryFilter(cmd *cobra.Command) string {
	queryStr, err := cmd.Root().PersistentFlags().GetString("query")
	if err != nil {
		queryStr = ""
	}
	output.SetOutputQuery(queryStr)
	return queryStr
}

// enforceQueryFormatGuard returns a usage error if --query is set and the
// resolved output format is not json. The caller must pass in the already-
// resolved format so this function does not re-derive it; this avoids a
// discrepancy when a subcommand defines a local --output flag that shadows
// the root persistent flag.
func enforceQueryFormatGuard(cmd *cobra.Command, queryStr, format string) error {
	if queryStr == "" {
		return nil
	}
	if format != FormatJSON {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return exitcodes.NewUsageError(fmt.Errorf("--query flag requires --output json"))
	}
	return nil
}

// applyFieldsFilter reads the --fields persistent flag and calls output.SetOutputFields.
// Called by all RunE wrappers so the filter is always consistent regardless of wrapper used.
func applyFieldsFilter(cmd *cobra.Command) {
	fieldsStr, err := cmd.Root().PersistentFlags().GetString("fields")
	if err != nil {
		fieldsStr = ""
	}
	if fieldsStr != "" {
		var fields []string
		for _, f := range strings.Split(fieldsStr, ",") {
			if f = strings.TrimSpace(f); f != "" {
				fields = append(fields, f)
			}
		}
		output.SetOutputFields(fields)
	} else {
		output.SetOutputFields(nil)
	}
}

// WrapRunE wraps a RunE function to set SilenceUsage to true if an error occurs and formats the error message.
func WrapRunE(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		applyFieldsFilter(cmd)
		format, _ := cmd.Root().PersistentFlags().GetString("output")
		if err := enforceQueryFormatGuard(cmd, applyQueryFilter(cmd), format); err != nil {
			return err
		}
		err := runE(cmd, args)
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			code := classifyError(err)
			wrapped := fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag", cmd.Name(), err, cmd.Name(), args)
			return exitcodes.New(code, wrapped)
		}
		return nil
	}
}

// WrapColorAwareRunE combines WrapRunE and WrapCommandFunc to handle both error formatting
// and passing the noColor flag to command functions.
func WrapColorAwareRunE(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		applyFieldsFilter(cmd)
		format, _ := cmd.Root().PersistentFlags().GetString("output")
		if err := enforceQueryFormatGuard(cmd, applyQueryFilter(cmd), format); err != nil {
			return err
		}
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
			code := classifyError(err)
			wrapped := fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag",
				cmd.Name(), err, cmd.Name(), args)
			return exitcodes.New(code, wrapped)
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
			return exitcodes.NewUsageError(fmt.Errorf("invalid output format: %s. Must be one of: %v", format, ValidFormats))
		}

		applyFieldsFilter(cmd)
		// Pass the already-resolved format so enforceQueryFormatGuard does not
		// re-read --output from the root persistent flags (which could differ
		// if the subcommand defines its own local --output override).
		if err := enforceQueryFormatGuard(cmd, applyQueryFilter(cmd), format); err != nil {
			return err
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
			code := classifyError(err)
			wrapped := fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag",
				cmd.Name(), err, cmd.Name(), args)
			return exitcodes.New(code, wrapped)
		}
		return nil
	}
}

// classifyError inspects an error message to determine the appropriate exit code.
func classifyError(err error) int {
	// Preserve exit codes already set by action functions
	var cliErr *exitcodes.CLIError
	if errors.As(err, &cliErr) {
		return cliErr.Code
	}

	// Type-safe SDK error inspection first
	var apiErr *megaport.ErrorResponse
	if errors.As(err, &apiErr) {
		switch apiErr.Response.StatusCode {
		case 401, 403:
			return exitcodes.Authentication
		case 404, 422, 429, 500, 502, 503:
			return exitcodes.API
		}
	}

	msg := err.Error()

	// Authentication errors
	authPatterns := []string{
		"error logging in",
		"access key not provided",
		"secret key not provided",
		"authentication",
		"Authorize",
	}
	for _, p := range authPatterns {
		if strings.Contains(msg, p) {
			return exitcodes.Authentication
		}
	}

	// Usage/validation errors
	usagePatterns := []string{
		"invalid output format",
		"required flag",
		"not set when not using interactive",
		"at least one field must be updated",
		"at least one of these flags",
		"invalid location ID",
	}
	for _, p := range usagePatterns {
		if strings.Contains(msg, p) {
			return exitcodes.Usage
		}
	}
	// "invalid" + "ID" pattern
	if strings.Contains(msg, "invalid") && strings.Contains(msg, "ID") {
		return exitcodes.Usage
	}

	// API errors
	apiPatterns := []string{
		"error listing",
		"error getting",
		"error creating",
		"error updating",
		"error deleting",
		"error buying",
		"error modifying",
		"failed to retrieve",
		"failed to buy",
		"failed to validate",
		"API failure",
	}
	for _, p := range apiPatterns {
		if strings.Contains(msg, p) {
			return exitcodes.API
		}
	}

	return exitcodes.General
}
