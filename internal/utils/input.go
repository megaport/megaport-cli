package utils

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
)

// InputConfig configures how ResolveInput resolves the request object from
// one of the three input modes (JSON, flags, interactive prompts).
type InputConfig[T any] struct {
	// ResourceName is used in the fallback error message (e.g. "port", "MCR").
	ResourceName string

	// Cmd is the cobra command whose flags are inspected.
	Cmd *cobra.Command

	// NoColor disables colored output.
	NoColor bool

	// FlagsProvided should return true when any resource-specific CLI flags
	// have been explicitly set by the user.
	FlagsProvided func() bool

	// FromJSON builds the request from --json or --json-file input.
	FromJSON func(jsonStr, jsonFile string) (T, error)

	// FromFlags builds the request from CLI flags.
	FromFlags func() (T, error)

	// FromPrompt builds the request via interactive prompts.
	// May be nil if the command does not support interactive mode.
	FromPrompt func() (T, error)
}

// ResolveInput determines which input mode the user chose (JSON, flags, or
// interactive) and delegates to the appropriate builder function.
//
// Precedence: JSON > flags > interactive > error.
func ResolveInput[T any](cfg InputConfig[T]) (T, error) {
	var zero T

	jsonStr, _ := cfg.Cmd.Flags().GetString("json")
	jsonFile, _ := cfg.Cmd.Flags().GetString("json-file")
	interactive, _ := cfg.Cmd.Flags().GetBool("interactive")

	switch {
	case jsonStr != "" || jsonFile != "":
		output.PrintInfo("Using JSON input", cfg.NoColor)
		result, err := cfg.FromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", cfg.NoColor, err)
			return zero, err
		}
		return result, nil

	case cfg.FlagsProvided != nil && cfg.FlagsProvided():
		output.PrintInfo("Using flag input", cfg.NoColor)
		result, err := cfg.FromFlags()
		if err != nil {
			output.PrintError("Failed to process flag input: %v", cfg.NoColor, err)
			return zero, err
		}
		return result, nil

	case interactive && cfg.FromPrompt != nil:
		output.PrintInfo("Starting interactive mode", cfg.NoColor)
		result, err := cfg.FromPrompt()
		if err != nil {
			output.PrintError("Interactive input failed: %v", cfg.NoColor, err)
			return zero, err
		}
		return result, nil

	default:
		output.PrintError("No input provided", cfg.NoColor)
		return zero, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify %s details", cfg.ResourceName)
	}
}
