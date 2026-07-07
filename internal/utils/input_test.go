package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Bool("interactive", false, "")
	return cmd
}

func TestResolveInput_JSON(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("json", `{"name":"test"}`))

	result, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON: func(jsonStr, jsonFile string) (string, error) {
			return "from-json", nil
		},
		FromFlags:  func() (string, error) { return "", fmt.Errorf("should not be called") },
		FromPrompt: func() (string, error) { return "", fmt.Errorf("should not be called") },
	})
	assert.NoError(t, err)
	assert.Equal(t, "from-json", result)
}

func TestResolveInput_Flags(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test"))

	result, err := ResolveInput(InputConfig[string]{
		ResourceName:  "port",
		Cmd:           cmd,
		NoColor:       true,
		FlagsProvided: func() bool { return cmd.Flags().Changed("name") },
		FromJSON:      func(jsonStr, jsonFile string) (string, error) { return "", fmt.Errorf("should not be called") },
		FromFlags:     func() (string, error) { return "from-flags", nil },
		FromPrompt:    func() (string, error) { return "", fmt.Errorf("should not be called") },
	})
	assert.NoError(t, err)
	assert.Equal(t, "from-flags", result)
}

func TestResolveInput_Interactive(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("interactive", "true"))

	result, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", fmt.Errorf("should not be called") },
		FromFlags:    func() (string, error) { return "", fmt.Errorf("should not be called") },
		FromPrompt:   func() (string, error) { return "from-prompt", nil },
	})
	assert.NoError(t, err)
	assert.Equal(t, "from-prompt", result)
}

func TestResolveInput_NoInput(t *testing.T) {
	cmd := newTestCmd()

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", nil },
		FromFlags:    func() (string, error) { return "", nil },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
	assert.Contains(t, err.Error(), "port")
}

func TestResolveInput_JSONPrecedenceOverFlags(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.Flags().Set("json", `{}`))
	require.NoError(t, cmd.Flags().Set("name", "test"))

	result, err := ResolveInput(InputConfig[string]{
		ResourceName:  "port",
		Cmd:           cmd,
		NoColor:       true,
		FlagsProvided: func() bool { return cmd.Flags().Changed("name") },
		FromJSON:      func(jsonStr, jsonFile string) (string, error) { return "from-json", nil },
		FromFlags:     func() (string, error) { return "from-flags", nil },
	})
	assert.NoError(t, err)
	assert.Equal(t, "from-json", result)
}

func TestResolveInput_JSONError(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("json", `invalid`))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", fmt.Errorf("parse error") },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestResolveInput_FlagsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test"))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName:  "port",
		Cmd:           cmd,
		NoColor:       true,
		FlagsProvided: func() bool { return true },
		FromJSON:      func(jsonStr, jsonFile string) (string, error) { return "", nil },
		FromFlags:     func() (string, error) { return "", fmt.Errorf("flag validation failed") },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flag validation failed")
}

func TestResolveInput_PromptError(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("interactive", "true"))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", nil },
		FromPrompt:   func() (string, error) { return "", fmt.Errorf("prompt cancelled") },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt cancelled")
}

func TestResolveInput_JSONFile(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("json-file", "/tmp/test.json"))

	result, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON: func(jsonStr, jsonFile string) (string, error) {
			assert.Equal(t, "/tmp/test.json", jsonFile)
			return "from-json-file", nil
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "from-json-file", result)
}

func TestResolveInput_NilFromPromptFallsToError(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("interactive", "true"))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "MCR",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", nil },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
	assert.Contains(t, err.Error(), "MCR")
}

func TestResolveInput_NilFromJSON(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("json", `{}`))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no JSON handler configured")
}

func TestResolveInput_NilFromFlags(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test"))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName:  "port",
		Cmd:           cmd,
		NoColor:       true,
		FlagsProvided: func() bool { return true },
		FromJSON:      func(jsonStr, jsonFile string) (string, error) { return "", nil },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no flag handler configured")
}

func TestResolveInput_NilCmd(t *testing.T) {
	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no command configured")
}

func TestResolveInput_InteractiveWithFlagsIsConflict(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.Flags().Set("interactive", "true"))
	require.NoError(t, cmd.Flags().Set("name", "test"))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName:  "port",
		Cmd:           cmd,
		NoColor:       true,
		FlagsProvided: func() bool { return cmd.Flags().Changed("name") },
		FromJSON:      func(jsonStr, jsonFile string) (string, error) { return "", fmt.Errorf("should not be called") },
		FromFlags:     func() (string, error) { return "", fmt.Errorf("should not be called") },
		FromPrompt:    func() (string, error) { return "", fmt.Errorf("should not be called") },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined with")
}

func TestResolveInput_InteractiveWithJSONIsConflict(t *testing.T) {
	cmd := newTestCmd()
	require.NoError(t, cmd.Flags().Set("interactive", "true"))
	require.NoError(t, cmd.Flags().Set("json", `{}`))

	_, err := ResolveInput(InputConfig[string]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      true,
		FromJSON:     func(jsonStr, jsonFile string) (string, error) { return "", fmt.Errorf("should not be called") },
		FromPrompt:   func() (string, error) { return "", fmt.Errorf("should not be called") },
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined with")
}

func TestCheckInteractiveConflict(t *testing.T) {
	assert.NoError(t, CheckInteractiveConflict(false, false))
	assert.NoError(t, CheckInteractiveConflict(false, true))
	assert.NoError(t, CheckInteractiveConflict(true, false))
	err := CheckInteractiveConflict(true, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined with")

	// The ticket mandates usage exit code 2, so assert it at the source rather
	// than only matching the message (which a plain fmt.Errorf would also pass).
	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr), "conflict error must be a CLIError")
	assert.Equal(t, exitcodes.Usage, cliErr.Code)
}

// newConflictTestCmd builds a leaf command wired to a parent that owns the
// global persistent flags, mirroring how real commands sit under the root. This
// lets the test prove inherited globals never trip the conflict detector.
func newConflictTestCmd() *cobra.Command {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "table", "")
	root.PersistentFlags().Bool("no-color", false, "")
	root.PersistentFlags().String("env", "production", "")

	leaf := &cobra.Command{Use: "buy"}
	leaf.Flags().Bool("interactive", false, "")
	leaf.Flags().Bool("generate-skeleton", false, "")
	leaf.Flags().Bool("yes", false, "")
	leaf.Flags().Bool("no-wait", false, "")
	leaf.Flags().Bool("force", false, "")
	leaf.Flags().Bool("export", false, "")
	leaf.Flags().String("json", "", "")
	leaf.Flags().String("json-file", "", "")
	leaf.Flags().String("name", "", "")
	leaf.Flags().String("cost-centre", "", "")
	root.AddCommand(leaf)
	return leaf
}

func TestHasConflictingInputFlags(t *testing.T) {
	tests := []struct {
		name     string
		set      map[string]string
		expected bool
	}{
		{name: "nothing set", set: nil, expected: false},
		{name: "only interactive", set: map[string]string{"interactive": "true"}, expected: false},
		{name: "behavior flags only", set: map[string]string{"interactive": "true", "yes": "true", "no-wait": "true", "force": "true", "export": "true", "generate-skeleton": "true"}, expected: false},
		{name: "value flag", set: map[string]string{"interactive": "true", "name": "foo"}, expected: true},
		{name: "optional value flag", set: map[string]string{"interactive": "true", "cost-centre": "eng"}, expected: true},
		{name: "json", set: map[string]string{"interactive": "true", "json": "{}"}, expected: true},
		{name: "json-file", set: map[string]string{"interactive": "true", "json-file": "/tmp/x.json"}, expected: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newConflictTestCmd()
			for k, v := range tt.set {
				require.NoError(t, cmd.Flags().Set(k, v))
			}
			assert.Equal(t, tt.expected, HasConflictingInputFlags(cmd))
		})
	}
}

func TestHasConflictingInputFlags_InheritedGlobalsIgnored(t *testing.T) {
	// Drive the real cobra parse path so inherited globals land in the merged
	// flagset exactly as they do at runtime, then assert they don't count as a
	// conflict. Isolated leaf.Flags().Set can't reach parent persistent flags.
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "table", "")
	root.PersistentFlags().Bool("no-color", false, "")

	var got, ran bool
	leaf := &cobra.Command{
		Use: "buy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ran = true
			got = HasConflictingInputFlags(cmd)
			return nil
		},
	}
	leaf.Flags().Bool("interactive", false, "")
	leaf.Flags().String("name", "", "")
	root.AddCommand(leaf)

	root.SetArgs([]string{"buy", "--interactive", "--output", "json", "--no-color"})
	require.NoError(t, root.Execute())
	require.True(t, ran, "leaf RunE did not run")
	assert.False(t, got, "inherited globals must not count as a conflict")
}

func TestHasConflictingInputFlags_ValueFlagWithGlobals(t *testing.T) {
	// Same real parse path, but now a value flag is set alongside the globals:
	// this must be detected as a conflict.
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "table", "")

	var got, ran bool
	leaf := &cobra.Command{
		Use: "buy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ran = true
			got = HasConflictingInputFlags(cmd)
			return nil
		},
	}
	leaf.Flags().Bool("interactive", false, "")
	leaf.Flags().String("name", "", "")
	root.AddCommand(leaf)

	root.SetArgs([]string{"buy", "--interactive", "--name", "foo", "--output", "json"})
	require.NoError(t, root.Execute())
	require.True(t, ran, "leaf RunE did not run")
	assert.True(t, got, "value flag alongside globals must count as a conflict")
}

func TestReadJSONInput_FromString(t *testing.T) {
	data, err := ReadJSONInput(`{"name":"test"}`, "")
	require.NoError(t, err)
	assert.Equal(t, `{"name":"test"}`, string(data))
}

func TestReadJSONInput_StringOverridesFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.json"
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"from":"file"}`), 0644))

	data, err := ReadJSONInput(`{"from":"string"}`, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, `{"from":"string"}`, string(data))
}
