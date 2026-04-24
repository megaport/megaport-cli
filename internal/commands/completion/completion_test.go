package completion

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		contains string
	}{
		{name: "bash", shell: "bash", contains: "__start_megaport-cli"},
		{name: "zsh", shell: "zsh", contains: "#compdef megaport-cli"},
		{name: "fish", shell: "fish", contains: "complete -c megaport-cli"},
		{name: "powershell", shell: "powershell", contains: "Register-ArgumentCompleter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "megaport-cli"}
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			AddCommandsTo(rootCmd)

			rootCmd.SetArgs([]string{"completion", tt.shell})
			err := rootCmd.Execute()

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}

func TestRunCompletion_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{name: "no args", args: []string{"completion"}, errMsg: "accepts 1 arg"},
		{name: "too many args", args: []string{"completion", "bash", "extra"}, errMsg: "accepts 1 arg"},
		{name: "invalid shell", args: []string{"completion", "invalid"}, errMsg: "invalid shell type"},
		{name: "empty arg", args: []string{"completion", ""}, errMsg: "invalid shell type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "megaport-cli", SilenceUsage: true, SilenceErrors: true}
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			AddCommandsTo(rootCmd)

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestModule(t *testing.T) {
	m := NewModule()
	assert.Equal(t, "completion", m.Name())

	rootCmd := &cobra.Command{Use: "megaport-cli"}
	m.RegisterCommands(rootCmd)

	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "completion" {
			found = true
			break
		}
	}
	assert.True(t, found, "completion command should be registered")
}
