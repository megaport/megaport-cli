package completion

import (
	"os"

	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the completion command and adds it to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create completion command using command builder
	completionCmd := cmdbuilder.NewCommand("completion", "Generate completion script").
		WithArgs(cobra.ExactArgs(1)).
		WithRunFunc(RunCompletion).
		WithValidArgs([]string{"bash", "zsh", "fish", "powershell"}).
		WithExample("megaport-cli completion bash > ~/.bash_completion.d/megaport-cli").
		WithExample("megaport-cli completion zsh > \"${fpath[1]}/_megaport-cli\"").
		WithExample("megaport-cli completion fish > ~/.config/fish/completions/megaport-cli.fish").
		WithExample("megaport-cli completion powershell > megaport-cli.ps1").
		WithImportantNote("Bash: source <(megaport-cli completion bash)").
		WithImportantNote("Zsh: You need to enable shell completion with 'autoload -U compinit; compinit'").
		WithImportantNote("Fish: megaport-cli completion fish | source").
		WithImportantNote("PowerShell: megaport-cli completion powershell | Out-String | Invoke-Expression").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(completionCmd)
}

// RunCompletion handles the completion command execution
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
