package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(megaport completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ megaport completion bash > /etc/bash_completion.d/megaport
  # macOS:
  $ megaport completion bash > /usr/local/etc/bash_completion.d/megaport

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ megaport completion zsh > "${fpath[1]}/_megaport"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ megaport completion fish | source

  # To load completions for each session, execute once:
  $ megaport completion fish > ~/.config/fish/completions/megaport.fish

PowerShell:

  PS> megaport completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> megaport completion powershell > megaport.ps1
  # and source this file from your PowerShell profile.
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(RunCompletion),
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
	rootCmd.AddCommand(completionCmd)
}
