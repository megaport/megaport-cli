package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCmd(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		wantErr  bool
		contains string
	}{
		{
			name:     "bash completion",
			shell:    "bash",
			wantErr:  false,
			contains: "__start_megaport",
		},
		{
			name:     "zsh completion",
			shell:    "zsh",
			wantErr:  false,
			contains: "#compdef",
		},
		{
			name:     "fish completion",
			shell:    "fish",
			wantErr:  false,
			contains: "complete -c megaport",
		},
		{
			name:     "powershell completion",
			shell:    "powershell",
			wantErr:  false,
			contains: "Register-ArgumentCompleter",
		},
		{
			name:    "invalid shell",
			shell:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "megaport"}
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			completionCmd := &cobra.Command{
				Use:       "completion [bash|zsh|fish|powershell]",
				ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
				Args:      cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					switch args[0] {
					case "bash":
						return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
					case "zsh":
						return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
					case "fish":
						return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
					case "powershell":
						return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
					default:
						return fmt.Errorf("invalid shell type %q", args[0])
					}
				},
			}
			rootCmd.AddCommand(completionCmd)

			rootCmd.SetArgs([]string{"completion", tt.shell})
			err := rootCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			output := buf.String()
			assert.NotEmpty(t, output)

			if tt.contains != "" {
				assert.Contains(t, output, tt.contains)
			}
		})
	}
}

func TestCompletionCmd_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no args",
			args:    []string{"completion"},
			wantErr: true,
			errMsg:  "accepts 1 arg",
		},
		{
			name:    "too many args",
			args:    []string{"completion", "bash", "extra"},
			wantErr: true,
			errMsg:  "accepts 1 arg",
		},
		{
			name:    "empty arg",
			args:    []string{"completion", ""},
			wantErr: true,
			errMsg:  "invalid shell type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "megaport"}

			completion := &cobra.Command{
				Use:          "completion [bash|zsh|fish|powershell]",
				Short:        "Generate completion script",
				ValidArgs:    []string{"bash", "zsh", "fish", "powershell"},
				Args:         cobra.ExactArgs(1),
				SilenceUsage: true,
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) != 1 {
						return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
					}
					if args[0] == "" {
						return fmt.Errorf("invalid shell type %q", args[0])
					}
					switch args[0] {
					case "bash":
						return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
					case "zsh":
						return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
					case "fish":
						return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
					case "powershell":
						return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
					default:
						return fmt.Errorf("invalid shell type %q", args[0])
					}
				},
			}
			cmd.AddCommand(completion)

			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)
			cmd.SetOut(outBuf)
			cmd.SetErr(errBuf)

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			assert.Error(t, err, "Expected error for case: %s", tt.name)
			assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text for case: %s", tt.name)
			assert.Empty(t, outBuf.String(), "No output expected for error case: %s", tt.name)
		})
	}
}
