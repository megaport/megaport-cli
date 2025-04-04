package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(RunCompletion),
}

func RunCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return cmd.Help()
	}
}

func init() {
	// Set up help builder for the completion command
	completionHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli completion",
		ShortDesc:   "Generate completion script",
		LongDesc:    "Generate shell completion scripts for Megaport CLI.\n\nThis command outputs shell completion code for various shell environments that can be used to enable tab-completion of Megaport CLI commands.",
		Examples: []string{
			"completion bash > ~/.bash_completion.d/megaport-cli",
			"completion zsh > \"${fpath[1]}/_megaport-cli\"",
			"completion fish > ~/.config/fish/completions/megaport-cli.fish",
			"completion powershell > megaport-cli.ps1",
		},
		ImportantNotes: []string{
			"Bash: source <(megaport-cli completion bash)",
			"Zsh: You need to enable shell completion with 'autoload -U compinit; compinit'",
			"Fish: megaport-cli completion fish | source",
			"PowerShell: megaport-cli completion powershell | Out-String | Invoke-Expression",
		},
	}
	completionCmd.Long = completionHelp.Build()

	rootCmd.AddCommand(completionCmd)
}
