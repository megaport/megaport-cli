package help

import "github.com/spf13/cobra"

// MarkFlagsRequired marks a list of flags as required for a command
func MarkFlagsRequired(cmd *cobra.Command, flags []string) error {
	for _, flag := range flags {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCommonBuyFlags adds common flags for buy commands
func AddCommonBuyFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	cmd.Flags().String("json", "", "JSON string containing configuration")
	cmd.Flags().String("json-file", "", "Path to JSON file containing configuration")
}

// AddCommonUpdateFlags adds common flags for update commands
func AddCommonUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	cmd.Flags().String("json", "", "JSON string containing configuration")
	cmd.Flags().String("json-file", "", "Path to JSON file containing configuration")
}
