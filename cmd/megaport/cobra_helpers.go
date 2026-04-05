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
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, subCmd := range cmd.Commands() {
		resetAllFlags(subCmd)
	}
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
