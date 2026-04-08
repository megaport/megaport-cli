package utils

import (
	"fmt"
	"os"
	"testing"

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

func TestReadJSONInput_FromString(t *testing.T) {
	data, err := ReadJSONInput(`{"name":"test"}`, "")
	require.NoError(t, err)
	assert.Equal(t, `{"name":"test"}`, string(data))
}

func TestReadJSONInput_FromFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.json"
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"key":"value"}`), 0644))

	data, err := ReadJSONInput("", tmpFile)
	require.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, string(data))
}

func TestReadJSONInput_FileNotFound(t *testing.T) {
	_, err := ReadJSONInput("", "/nonexistent/path.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read JSON file")
}

func TestReadJSONInput_StringOverridesFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.json"
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"from":"file"}`), 0644))

	data, err := ReadJSONInput(`{"from":"string"}`, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, `{"from":"string"}`, string(data))
}
