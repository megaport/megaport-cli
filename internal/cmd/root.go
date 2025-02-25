package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "megaport",
	Short: "A CLI tool to interact with the Megaport API",
	Long: `A CLI tool to interact with the Megaport API.

This CLI supports the following features:
  - Configure credentials: Use "megaport configure" to set your access and secret keys.
  - Locations: List and manage locations.
  - Ports: List all ports and get details for a specific port.
  - MCRs: Get details for Megaport Cloud Routers.
  - MVEs: Get details for Megaport Virtual Edge devices.
  - VXCs: Get details for Virtual Cross Connects.
  - Partner Ports: List and filter partner ports based on product name, connect type, company name, location ID, and diversity zone.
`,
}

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

Fish:
  $ megaport completion fish > ~/.config/fish/completions/megaport.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("invalid shell type %q", args[0])
		}
	},
}

const (
	formatTable = "table"
	formatJSON  = "json"
	formatCSV   = "csv"
	formatXML   = "xml"
)

var (
	env          string
	outputFormat string
	validFormats = []string{formatTable, formatJSON, formatCSV, formatXML}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", formatTable,
		fmt.Sprintf("Output format (%s)", strings.Join(validFormats, ", ")))

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		format := strings.ToLower(outputFormat)
		for _, validFormat := range validFormats {
			if format == validFormat {
				outputFormat = format
				return nil
			}
		}
		return fmt.Errorf("invalid output format: %s. Must be one of: %s",
			outputFormat, strings.Join(validFormats, ", "))
	}
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "production", "Environment to use (production, staging, development)")
	err := rootCmd.PersistentFlags().SetAnnotation("output", cobra.BashCompCustom, validFormats)
	if err != nil {
		fmt.Println(err)
	}
	rootCmd.AddCommand(completionCmd)
}

func initConfig() {
	// Any additional initialization can be done here
}
