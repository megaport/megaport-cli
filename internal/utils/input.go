package utils

import (
	"errors"
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ErrInteractiveConflict is returned when --interactive is combined with
// --json, --json-file, or other value flags. The input modes are mutually
// exclusive: silently picking one would ignore the other input the user
// explicitly provided. Shared with the builder's required-flag gating
// (cmdbuilder.WithConditionalRequirements) so both surfaces agree.
var ErrInteractiveConflict = errors.New("--interactive cannot be combined with --json, --json-file, or other flags; choose a single input mode")

// CheckInteractiveConflict rejects --interactive combined with --json,
// --json-file, or other value flags. Update commands typically apply partial
// changes via a hand-rolled JSON/flags/interactive chain rather than going
// through ResolveInput; call this first in that chain so they enforce the
// same mutual-exclusivity policy ResolveInput does. Pass HasConflictingInputFlags(cmd)
// as hasOtherInput so every command detects the conflict the same way.
func CheckInteractiveConflict(interactive, hasOtherInput bool) error {
	if interactive && hasOtherInput {
		return exitcodes.NewUsageError(ErrInteractiveConflict)
	}
	return nil
}

// nonInputFlags are command-local flags that don't supply resource input, so
// setting one alongside --interactive is not a conflict: the input-mode
// selectors themselves plus behavior toggles (confirmation skips, wait
// control, output-only export). Global/persistent flags (--output, --env, ...)
// are excluded structurally by HasConflictingInputFlags and need not be listed.
var nonInputFlags = map[string]bool{
	"interactive":       true,
	"generate-skeleton": true,
	"yes":               true,
	"no-wait":           true,
	"force":             true,
	"export":            true,
}

// HasConflictingInputFlags reports whether the user explicitly set any
// command-local flag that supplies resource input (--json, --json-file, or a
// value flag) and therefore conflicts with --interactive. Inherited
// global/persistent flags (--output, --no-color, --env, --timeout, ...) never
// count, and behavior toggles are excluded via nonInputFlags. This keeps every
// buy/update/create command's conflict detection identical and drift-free as
// flags are added.
//
// We iterate cmd.Flags().Visit (the merged flagset's changed flags) rather than
// LocalNonPersistentFlags().Visit: cobra rebuilds the latter with AddFlag, which
// populates the set's formal map but not its actual map, so Visit sees nothing.
// Lookup, by contrast, reads formal and works on the rebuilt inherited set.
func HasConflictingInputFlags(cmd *cobra.Command) bool {
	inherited := cmd.InheritedFlags()
	conflict := false
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if nonInputFlags[f.Name] {
			return
		}
		if inherited.Lookup(f.Name) != nil {
			return
		}
		conflict = true
	})
	return conflict
}

// ReadJSONInput reads JSON data from either a raw string or a file path.
// If jsonStr is non-empty, it takes precedence; otherwise reads from jsonFile.
func ReadJSONInput(jsonStr, jsonFile string) ([]byte, error) {
	if jsonStr != "" {
		return []byte(jsonStr), nil
	}
	if jsonFile != "" {
		return readInputFile(jsonFile)
	}
	return nil, nil
}

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
// Combining --interactive with --json/--json-file or other value flags is a
// usage error (see ErrInteractiveConflict): the modes are mutually exclusive,
// so ResolveInput never silently drops one in favor of another. Otherwise
// precedence is JSON > flags > interactive > error.
func ResolveInput[T any](cfg InputConfig[T]) (T, error) {
	var zero T

	if cfg.Cmd == nil {
		return zero, fmt.Errorf("no command configured for %s input resolution", cfg.ResourceName)
	}

	jsonStr, err := cfg.Cmd.Flags().GetString("json")
	if err != nil {
		return zero, fmt.Errorf("failed to read --json flag: %w", err)
	}
	jsonFile, err := cfg.Cmd.Flags().GetString("json-file")
	if err != nil {
		return zero, fmt.Errorf("failed to read --json-file flag: %w", err)
	}
	interactive, err := cfg.Cmd.Flags().GetBool("interactive")
	if err != nil {
		return zero, fmt.Errorf("failed to read --interactive flag: %w", err)
	}

	hasJSON := jsonStr != "" || jsonFile != ""
	hasFlags := cfg.FlagsProvided != nil && cfg.FlagsProvided()

	// Detect the conflict from the full command-local flag set, not just the
	// primary flags used for mode selection below, so optional value flags
	// (--cost-centre, --promo-code, --resource-tags, ...) can't slip past.
	if err := CheckInteractiveConflict(interactive, HasConflictingInputFlags(cfg.Cmd)); err != nil {
		output.PrintError("%v", cfg.NoColor, err)
		return zero, err
	}

	switch {
	case hasJSON:
		if cfg.FromJSON == nil {
			return zero, fmt.Errorf("JSON input provided but no JSON handler configured for %s", cfg.ResourceName)
		}
		output.PrintInfo("Using JSON input", cfg.NoColor)
		result, err := cfg.FromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", cfg.NoColor, err)
			return zero, err
		}
		return result, nil

	case hasFlags:
		if cfg.FromFlags == nil {
			return zero, fmt.Errorf("flag input provided but no flag handler configured for %s", cfg.ResourceName)
		}
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
