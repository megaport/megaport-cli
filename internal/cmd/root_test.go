package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupCompletionTest() (*cobra.Command, *bytes.Buffer) {
	// Create a new root command
	cmd := &cobra.Command{Use: "megaport"}

	// Add the completion command
	completion := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate completion script",
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
	cmd.AddCommand(completion)

	// Setup output buffer
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	return cmd, buf
}

func TestCompletionCmd(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		wantErr  bool
		contains []string
	}{
		{
			name:    "bash completion",
			shell:   "bash",
			wantErr: false,
			contains: []string{
				"# bash completion for megaport",
				"__megaport_handle_word",
			},
		},
		{
			name:    "zsh completion",
			shell:   "zsh",
			wantErr: false,
			contains: []string{
				"#compdef megaport",
				"compdef _megaport megaport",
			},
		},
		{
			name:    "fish completion",
			shell:   "fish",
			wantErr: false,
			contains: []string{
				"complete -c megaport",
			},
		},
		{
			name:    "powershell completion",
			shell:   "powershell",
			wantErr: false,
			contains: []string{
				"Register-ArgumentCompleter",
				"CommandName 'megaport'",
			},
		},
		{
			name:    "invalid shell",
			shell:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, buf := setupCompletionTest()
			cmd.SetArgs([]string{"completion", tt.shell})
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			output := buf.String()

			assert.NotEmpty(t, output, "Expected non-empty completion script")
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected,
					"Output should contain expected content for %s completion", tt.shell)
			}
		})
	}
}
