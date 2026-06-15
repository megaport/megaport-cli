package megaport

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetAllFlags resets the value and "changed" state of every flag on every
// command in the tree. Without this, flags like --country set in a previous
// WASM invocation would appear as "Changed" in subsequent invocations even
// when the user did not supply them.
func resetAllFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(resetFlag)
	cmd.PersistentFlags().VisitAll(resetFlag)
	for _, subCmd := range cmd.Commands() {
		resetAllFlags(subCmd)
	}
}

// resetFlag restores a single flag to its default. Slice/array flags need
// Replace: their DefValue stringifies to "[]" and Set appends, so re-Setting
// DefValue would inject a literal "[]" element instead of clearing them.
func resetFlag(f *pflag.Flag) {
	if sv, ok := f.Value.(pflag.SliceValue); ok {
		_ = sv.Replace([]string{})
	} else {
		_ = f.Value.Set(f.DefValue)
	}
	f.Changed = false
}

// enableTraversalForAllCommands enables subcommand traversal on all commands
// in the tree recursively.
func enableTraversalForAllCommands(cmd *cobra.Command) {
	cmd.TraverseChildren = true
	cmd.Flags().ParseErrorsAllowlist.UnknownFlags = true

	for _, subCmd := range cmd.Commands() {
		enableTraversalForAllCommands(subCmd)
	}
}
