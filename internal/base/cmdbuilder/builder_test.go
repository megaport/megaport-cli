package cmdbuilder

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name      string
		use       string
		short     string
		wantUse   string
		wantShort string
	}{
		{
			name:      "basic command",
			use:       "list",
			short:     "List resources",
			wantUse:   "list",
			wantShort: "List resources",
		},
		{
			name:      "command with args in use",
			use:       "get [id]",
			short:     "Get a resource",
			wantUse:   "get [id]",
			wantShort: "Get a resource",
		},
		{
			name:      "empty strings",
			use:       "",
			short:     "",
			wantUse:   "",
			wantShort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewCommand(tt.use, tt.short)
			require.NotNil(t, b)
			require.NotNil(t, b.cmd)
			assert.Equal(t, tt.wantUse, b.cmd.Use)
			assert.Equal(t, tt.wantShort, b.cmd.Short)
			assert.NotNil(t, b.requiredFlags)
			assert.NotNil(t, b.optionalFlags)
			assert.NotNil(t, b.examples)
			assert.NotNil(t, b.importantNotes)
			assert.NotNil(t, b.jsonExamples)
		})
	}
}

func TestWithArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      cobra.PositionalArgs
		input     []string
		wantError bool
	}{
		{
			name:      "exact args passes",
			args:      cobra.ExactArgs(1),
			input:     []string{"arg1"},
			wantError: false,
		},
		{
			name:      "exact args fails with wrong count",
			args:      cobra.ExactArgs(1),
			input:     []string{},
			wantError: true,
		},
		{
			name:      "no args passes with empty",
			args:      cobra.NoArgs,
			input:     []string{},
			wantError: false,
		},
		{
			name:      "no args fails with args",
			args:      cobra.NoArgs,
			input:     []string{"arg1"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", "test").WithArgs(tt.args).Build()
			err := cmd.Args(cmd, tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithLongDesc(t *testing.T) {
	tests := []struct {
		name string
		desc string
	}{
		{name: "simple description", desc: "A long description"},
		{name: "multiline", desc: "Line one\nLine two\nLine three"},
		{name: "empty", desc: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewCommand("test", "test").WithLongDesc(tt.desc)
			assert.Equal(t, tt.desc, b.cmd.Long)
		})
	}
}

func TestWithRunFunc(t *testing.T) {
	called := false
	runFunc := func(cmd *cobra.Command, args []string) error {
		called = true
		return nil
	}

	cmd := NewCommand("test", "test").WithRunFunc(runFunc).Build()
	assert.NotNil(t, cmd.RunE)

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, called)

	// Test with error return
	errFunc := func(cmd *cobra.Command, args []string) error {
		return errors.New("test error")
	}
	cmd2 := NewCommand("test", "test").WithRunFunc(errFunc).Build()
	err = cmd2.RunE(cmd2, []string{})
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

func TestWithFlag(t *testing.T) {
	tests := []struct {
		name       string
		flagName   string
		defaultVal string
		usage      string
	}{
		{name: "basic flag", flagName: "name", defaultVal: "", usage: "The name"},
		{name: "flag with default", flagName: "format", defaultVal: "json", usage: "Output format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", "test").
				WithFlag(tt.flagName, tt.defaultVal, tt.usage).
				Build()

			f := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, f, "flag %q should exist", tt.flagName)
			assert.Equal(t, tt.defaultVal, f.DefValue)
			assert.Equal(t, tt.usage, f.Usage)
			assert.Equal(t, "string", f.Value.Type())
		})
	}
}

func TestWithFlagP(t *testing.T) {
	cmd := NewCommand("test", "test").
		WithFlagP("output", "o", "table", "Output format").
		Build()

	f := cmd.Flags().Lookup("output")
	require.NotNil(t, f)
	assert.Equal(t, "table", f.DefValue)
	assert.Equal(t, "o", f.Shorthand)
	assert.Equal(t, "string", f.Value.Type())
}

func TestWithIntFlag(t *testing.T) {
	tests := []struct {
		name       string
		flagName   string
		defaultVal int
		usage      string
	}{
		{name: "zero default", flagName: "count", defaultVal: 0, usage: "Count"},
		{name: "non-zero default", flagName: "port", defaultVal: 8080, usage: "Port number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", "test").
				WithIntFlag(tt.flagName, tt.defaultVal, tt.usage).
				Build()

			f := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, f, "flag %q should exist", tt.flagName)
			assert.Equal(t, "int", f.Value.Type())
		})
	}
}

func TestWithIntFlagP(t *testing.T) {
	cmd := NewCommand("test", "test").
		WithIntFlagP("term", "t", 12, "Contract term").
		Build()

	f := cmd.Flags().Lookup("term")
	require.NotNil(t, f)
	assert.Equal(t, "12", f.DefValue)
	assert.Equal(t, "t", f.Shorthand)
	assert.Equal(t, "int", f.Value.Type())
}

func TestWithBoolFlag(t *testing.T) {
	tests := []struct {
		name       string
		flagName   string
		defaultVal bool
	}{
		{name: "default false", flagName: "verbose", defaultVal: false},
		{name: "default true", flagName: "enabled", defaultVal: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", "test").
				WithBoolFlag(tt.flagName, tt.defaultVal, "test flag").
				Build()

			f := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, f, "flag %q should exist", tt.flagName)
			assert.Equal(t, "bool", f.Value.Type())
			if tt.defaultVal {
				assert.Equal(t, "true", f.DefValue)
			} else {
				assert.Equal(t, "false", f.DefValue)
			}
		})
	}
}

func TestWithBoolFlagP(t *testing.T) {
	cmd := NewCommand("test", "test").
		WithBoolFlagP("interactive", "i", false, "Interactive mode").
		Build()

	f := cmd.Flags().Lookup("interactive")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
	assert.Equal(t, "i", f.Shorthand)
	assert.Equal(t, "bool", f.Value.Type())
}

func TestWithDurationFlag(t *testing.T) {
	cmd := NewCommand("test", "test").
		WithDurationFlag("timeout", 30*time.Second, "Timeout duration").
		Build()

	f := cmd.Flags().Lookup("timeout")
	require.NotNil(t, f)
	assert.Equal(t, "30s", f.DefValue)
	assert.Equal(t, "duration", f.Value.Type())
}

func TestWithDurationFlagP(t *testing.T) {
	cmd := NewCommand("test", "test").
		WithDurationFlagP("interval", "n", 5*time.Second, "Polling interval").
		Build()

	f := cmd.Flags().Lookup("interval")
	require.NotNil(t, f)
	assert.Equal(t, "5s", f.DefValue)
	assert.Equal(t, "n", f.Shorthand)
	assert.Equal(t, "duration", f.Value.Type())
}

func TestWithRequiredFlag(t *testing.T) {
	t.Run("existing flag", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithFlag("name", "", "The name").
			WithRequiredFlag("name", "Name of the resource").
			Build()

		f := cmd.Flags().Lookup("name")
		require.NotNil(t, f)
		assert.Contains(t, f.Usage, "[required]")
		assert.Equal(t, "Name of the resource [required]", f.Usage)
		assert.NotNil(t, f.Annotations)
		assert.Equal(t, []string{"true"}, f.Annotations["cobra_annotation_bash_completion_one_required_flag"])
	})

	t.Run("non-existing flag does not panic", func(t *testing.T) {
		b := NewCommand("test", "test").
			WithRequiredFlag("nonexistent", "Does not exist")
		// Should not panic, just store in requiredFlags map
		assert.Equal(t, "Does not exist", b.requiredFlags["nonexistent"])
	})
}

func TestWithDocumentedRequiredFlag(t *testing.T) {
	t.Run("existing flag", func(t *testing.T) {
		b := NewCommand("test", "test").
			WithFlag("name", "", "The name").
			WithDocumentedRequiredFlag("name", "Name of the resource")

		f := b.cmd.Flags().Lookup("name")
		require.NotNil(t, f)
		assert.Equal(t, "Name of the resource [required]", f.Usage)
		// Should NOT have cobra annotation (unlike WithRequiredFlag)
		assert.Nil(t, f.Annotations)
		// Should be stored in requiredFlags
		assert.Equal(t, "Name of the resource", b.requiredFlags["name"])
	})

	t.Run("non-existing flag does not panic", func(t *testing.T) {
		b := NewCommand("test", "test").
			WithDocumentedRequiredFlag("nonexistent", "Does not exist")
		assert.Equal(t, "Does not exist", b.requiredFlags["nonexistent"])
	})
}

func TestWithConditionalRequirements(t *testing.T) {
	buildTestCmd := func() *cobra.Command {
		return NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "Interactive mode").
			WithFlag("json", "", "JSON input").
			WithFlag("json-file", "", "JSON file path").
			WithFlag("name", "", "Name").
			WithFlag("location", "", "Location").
			WithConditionalRequirements("name").
			Build()
	}

	t.Run("fails when required flag not set", func(t *testing.T) {
		cmd := buildTestCmd()
		err := cmd.PreRunE(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("passes when required flag is set", func(t *testing.T) {
		cmd := buildTestCmd()
		require.NoError(t, cmd.Flags().Set("name", "test-name"))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("passes when interactive is set", func(t *testing.T) {
		cmd := buildTestCmd()
		require.NoError(t, cmd.Flags().Set("interactive", "true"))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("passes when json is set", func(t *testing.T) {
		cmd := buildTestCmd()
		require.NoError(t, cmd.Flags().Set("json", `{"name":"test"}`))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("passes when json-file is set", func(t *testing.T) {
		cmd := buildTestCmd()
		require.NoError(t, cmd.Flags().Set("json-file", "/tmp/config.json"))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("at_least_one prefix - fails when none set", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "").
			WithFlag("json", "", "").
			WithFlag("json-file", "", "").
			WithFlag("name", "", "").
			WithFlag("location", "", "").
			WithConditionalRequirements("at_least_one:name,location").
			Build()

		err := cmd.PreRunE(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one")
	})

	t.Run("at_least_one prefix - passes when one is set", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "").
			WithFlag("json", "", "").
			WithFlag("json-file", "", "").
			WithFlag("name", "", "").
			WithFlag("location", "", "").
			WithConditionalRequirements("at_least_one:name,location").
			Build()

		require.NoError(t, cmd.Flags().Set("location", "Sydney"))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("multiple conditional requirements", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "").
			WithFlag("json", "", "").
			WithFlag("json-file", "", "").
			WithFlag("name", "", "").
			WithFlag("location", "", "").
			WithConditionalRequirements("name", "location").
			Build()

		// Fails when only one is set
		require.NoError(t, cmd.Flags().Set("name", "test"))
		err := cmd.PreRunE(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location")
	})

	t.Run("chains with existing PreRunE", func(t *testing.T) {
		preRunCalled := false
		b := NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "").
			WithFlag("json", "", "").
			WithFlag("json-file", "", "").
			WithFlag("name", "", "")

		// Set a PreRunE manually before chaining
		b.cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			preRunCalled = true
			return nil
		}

		cmd := b.WithConditionalRequirements("name").Build()
		require.NoError(t, cmd.Flags().Set("name", "test"))
		err := cmd.PreRunE(cmd, []string{})
		assert.NoError(t, err)
		assert.True(t, preRunCalled)
	})

	t.Run("chains with failing existing PreRunE", func(t *testing.T) {
		b := NewCommand("test", "test").
			WithBoolFlagP("interactive", "i", false, "").
			WithFlag("json", "", "").
			WithFlag("json-file", "", "").
			WithFlag("name", "", "")

		b.cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return errors.New("pre-run failed")
		}

		cmd := b.WithConditionalRequirements("name").Build()
		require.NoError(t, cmd.Flags().Set("name", "test"))
		err := cmd.PreRunE(cmd, []string{})
		assert.Error(t, err)
		assert.Equal(t, "pre-run failed", err.Error())
	})
}

func TestWithOptionalFlag(t *testing.T) {
	b := NewCommand("test", "test").
		WithOptionalFlag("format", "Output format").
		WithOptionalFlag("verbose", "Enable verbose output")

	assert.Equal(t, "Output format", b.optionalFlags["format"])
	assert.Equal(t, "Enable verbose output", b.optionalFlags["verbose"])
}

func TestWithExample(t *testing.T) {
	b := NewCommand("test", "test").
		WithExample("megaport-cli test --name foo").
		WithExample("megaport-cli test --name bar --verbose")

	assert.Len(t, b.examples, 2)
	assert.Equal(t, "megaport-cli test --name foo", b.examples[0])
	assert.Equal(t, "megaport-cli test --name bar --verbose", b.examples[1])
}

func TestWithImportantNote(t *testing.T) {
	b := NewCommand("test", "test").
		WithImportantNote("This is note 1").
		WithImportantNote("This is note 2")

	assert.Len(t, b.importantNotes, 2)
	assert.Equal(t, "This is note 1", b.importantNotes[0])
	assert.Equal(t, "This is note 2", b.importantNotes[1])
}

func TestWithJSONExample(t *testing.T) {
	b := NewCommand("test", "test").
		WithJSONExample(`{"name": "test"}`).
		WithJSONExample(`{"name": "test2", "term": 12}`)

	assert.Len(t, b.jsonExamples, 2)
	assert.Equal(t, `{"name": "test"}`, b.jsonExamples[0])
	assert.Equal(t, `{"name": "test2", "term": 12}`, b.jsonExamples[1])
}

func TestWithRootCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "megaport-cli", Short: "CLI tool"}
	b := NewCommand("test", "test").WithRootCmd(rootCmd)
	assert.Equal(t, rootCmd, b.rootCmd)
}

func TestBuild(t *testing.T) {
	t.Run("basic build adds docs subcommand", func(t *testing.T) {
		cmd := NewCommand("test", "test command").Build()

		assert.Equal(t, "test", cmd.Use)
		assert.Equal(t, "test command", cmd.Short)

		// Verify docs subcommand exists
		docsCmd, _, err := cmd.Find([]string{"docs"})
		assert.NoError(t, err)
		assert.Equal(t, "docs", docsCmd.Use)
	})

	t.Run("build with root cmd sets long desc", func(t *testing.T) {
		rootCmd := &cobra.Command{Use: "megaport-cli"}
		cmd := NewCommand("test", "test command").
			WithLongDesc("A long description").
			WithRootCmd(rootCmd).
			Build()

		// Long should be rebuilt by the help builder
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("build preserves all settings", func(t *testing.T) {
		called := false
		cmd := NewCommand("create", "Create resource").
			WithLongDesc("Create a new resource").
			WithFlag("name", "", "Resource name").
			WithIntFlag("count", 1, "Count").
			WithBoolFlag("force", false, "Force").
			WithRunFunc(func(cmd *cobra.Command, args []string) error {
				called = true
				return nil
			}).
			WithValidArgs([]string{"port", "vxc"}).
			Build()

		assert.Equal(t, "create", cmd.Use)
		assert.Equal(t, "Create resource", cmd.Short)
		assert.NotNil(t, cmd.Flags().Lookup("name"))
		assert.NotNil(t, cmd.Flags().Lookup("count"))
		assert.NotNil(t, cmd.Flags().Lookup("force"))
		assert.Equal(t, []string{"port", "vxc"}, cmd.ValidArgs)
		assert.NotNil(t, cmd.RunE)

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWithValidArgs(t *testing.T) {
	tests := []struct {
		name      string
		validArgs []string
	}{
		{name: "single arg", validArgs: []string{"port"}},
		{name: "multiple args", validArgs: []string{"port", "vxc", "mcr"}},
		{name: "empty", validArgs: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", "test").WithValidArgs(tt.validArgs).Build()
			assert.Equal(t, tt.validArgs, cmd.ValidArgs)
		})
	}
}

func TestReflagCmd(t *testing.T) {
	t.Run("marks existing flags as required", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithFlag("name", "", "Name").
			WithIntFlag("id", 0, "ID").
			ReflagCmd("name", "id").
			Build()

		// Cobra marks required flags via annotations
		nameFlag := cmd.Flags().Lookup("name")
		require.NotNil(t, nameFlag)
		assert.Contains(t, nameFlag.Annotations, cobra.BashCompOneRequiredFlag)

		idFlag := cmd.Flags().Lookup("id")
		require.NotNil(t, idFlag)
		assert.Contains(t, idFlag.Annotations, cobra.BashCompOneRequiredFlag)
	})

	t.Run("non-existing flag prints warning but does not panic", func(t *testing.T) {
		// Should not panic even if flag doesn't exist
		assert.NotPanics(t, func() {
			NewCommand("test", "test").
				ReflagCmd("nonexistent").
				Build()
		})
	})
}

func TestWithValidArgsFunction(t *testing.T) {
	completionFunc := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"port", "vxc"}, cobra.ShellCompDirectiveNoFileComp
	}

	cmd := NewCommand("test", "test").
		WithValidArgsFunction(completionFunc).
		Build()

	assert.NotNil(t, cmd.ValidArgsFunction)
	completions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")
	assert.Equal(t, []string{"port", "vxc"}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestWithFlagCompletionFunc(t *testing.T) {
	t.Run("existing flag", func(t *testing.T) {
		cmd := NewCommand("test", "test").
			WithFlag("format", "table", "Output format").
			WithFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return []string{"table", "json", "csv"}, cobra.ShellCompDirectiveNoFileComp
			}).
			Build()

		assert.NotNil(t, cmd.Flags().Lookup("format"))
	})

	t.Run("non-existing flag does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			NewCommand("test", "test").
				WithFlagCompletionFunc("nonexistent", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					return nil, cobra.ShellCompDirectiveDefault
				}).
				Build()
		})
	})
}

func TestBuilderChaining(t *testing.T) {
	// Verify every builder method returns the same builder for chaining
	b := NewCommand("test", "test")
	assert.Equal(t, b, b.WithLongDesc("desc"))
	assert.Equal(t, b, b.WithFlag("f1", "", ""))
	assert.Equal(t, b, b.WithFlagP("f2", "x", "", ""))
	assert.Equal(t, b, b.WithIntFlag("f3", 0, ""))
	assert.Equal(t, b, b.WithIntFlagP("f4", "y", 0, ""))
	assert.Equal(t, b, b.WithBoolFlag("f5", false, ""))
	assert.Equal(t, b, b.WithBoolFlagP("f6", "z", false, ""))
	assert.Equal(t, b, b.WithDurationFlag("f7", time.Second, ""))
	assert.Equal(t, b, b.WithDurationFlagP("f8", "d", time.Second, ""))
	assert.Equal(t, b, b.WithExample("example"))
	assert.Equal(t, b, b.WithImportantNote("note"))
	assert.Equal(t, b, b.WithJSONExample("{}"))
	assert.Equal(t, b, b.WithOptionalFlag("opt", "desc"))
	assert.Equal(t, b, b.WithValidArgs([]string{"a"}))
	assert.Equal(t, b, b.WithRootCmd(&cobra.Command{}))
	assert.Equal(t, b, b.WithAliases([]string{"t"}))
}
