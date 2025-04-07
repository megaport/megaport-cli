package cmdbuilder

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

type CommandBuilder struct {
	cmd            *cobra.Command
	requiredFlags  map[string]string
	optionalFlags  map[string]string
	examples       []string
	importantNotes []string
	jsonExamples   []string
	rootCmd        *cobra.Command
}

// NewCommand creates a new command builder with minimal required fields
func NewCommand(use, short string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &cobra.Command{
			Use:   use,
			Short: short,
		},
		requiredFlags:  make(map[string]string),
		optionalFlags:  make(map[string]string),
		examples:       []string{},
		importantNotes: []string{},
		jsonExamples:   []string{},
	}
}

// WithArgs sets the positional arguments validator for the command
func (b *CommandBuilder) WithArgs(args cobra.PositionalArgs) *CommandBuilder {
	b.cmd.Args = args
	return b
}

// WithLongDesc sets the long description for the command
func (b *CommandBuilder) WithLongDesc(desc string) *CommandBuilder {
	b.cmd.Long = desc
	return b
}

// WithRunFunc sets the command's run function
func (b *CommandBuilder) WithRunFunc(f func(*cobra.Command, []string) error) *CommandBuilder {
	b.cmd.RunE = f
	return b
}

// WithOutputFormatRunFunc wraps the run function with output formatting
func (b *CommandBuilder) WithOutputFormatRunFunc(f func(*cobra.Command, []string, bool, string) error) *CommandBuilder {
	b.cmd.RunE = utils.WrapOutputFormatRunE(f)
	return b
}

// WithColorAwareRunFunc wraps the run function with color awareness
func (b *CommandBuilder) WithColorAwareRunFunc(f func(*cobra.Command, []string, bool) error) *CommandBuilder {
	b.cmd.RunE = utils.WrapColorAwareRunE(f)
	return b
}

// WithFlag adds a standard string flag to the command
func (b *CommandBuilder) WithFlag(name, defaultVal, usage string) *CommandBuilder {
	b.cmd.Flags().String(name, defaultVal, usage)
	return b
}

// WithFlagP adds a standard string flag with a shorthand to the command
func (b *CommandBuilder) WithFlagP(name, shorthand, defaultVal, usage string) *CommandBuilder {
	b.cmd.Flags().StringP(name, shorthand, defaultVal, usage)
	return b
}

// WithIntFlag adds an integer flag to the command
func (b *CommandBuilder) WithIntFlag(name string, defaultVal int, usage string) *CommandBuilder {
	b.cmd.Flags().Int(name, defaultVal, usage)
	return b
}

// WithIntFlagP adds an integer flag with a shorthand to the command
func (b *CommandBuilder) WithIntFlagP(name, shorthand string, defaultVal int, usage string) *CommandBuilder {
	b.cmd.Flags().IntP(name, shorthand, defaultVal, usage)
	return b
}

// WithBoolFlag adds a boolean flag to the command
func (b *CommandBuilder) WithBoolFlag(name string, defaultVal bool, usage string) *CommandBuilder {
	b.cmd.Flags().Bool(name, defaultVal, usage)
	return b
}

// WithBoolFlag adds a boolean flag to the command
func (b *CommandBuilder) WithBoolFlagP(name, shorthand string, defaultVal bool, usage string) *CommandBuilder {
	b.cmd.Flags().BoolP(name, shorthand, defaultVal, usage)
	return b
}

// WithRootCmd sets the root command for help generation
func (b *CommandBuilder) WithRootCmd(rootCmd *cobra.Command) *CommandBuilder {
	b.rootCmd = rootCmd
	return b
}

// WithRequiredFlag marks a flag as required and adds description
func (b *CommandBuilder) WithRequiredFlag(name, description string) *CommandBuilder {
	// Mark the flag in the cobra command
	if b.cmd.Flags().Lookup(name) != nil {
		flag := b.cmd.Flags().Lookup(name)

		// Add annotation for Cobra's bash completion
		if flag.Annotations == nil {
			flag.Annotations = make(map[string][]string)
		}
		flag.Annotations["cobra_annotation_bash_completion_one_required_flag"] = []string{"true"}

		// Also update the description to indicate it's required
		flag.Usage = description + " [required]"
	}

	// Store for our documentation as well
	b.requiredFlags[name] = description
	return b
}

// WithOptionalFlag adds documentation for an optional flag
func (b *CommandBuilder) WithOptionalFlag(name, desc string) *CommandBuilder {
	b.optionalFlags[name] = desc
	return b
}

// WithExample adds an example to the command's documentation
func (b *CommandBuilder) WithExample(example string) *CommandBuilder {
	b.examples = append(b.examples, example)
	return b
}

// WithJsonExample adds a JSON example to the command's documentation
func (b *CommandBuilder) WithJSONExample(example string) *CommandBuilder {
	b.jsonExamples = append(b.jsonExamples, example)
	return b
}

// WithImportantNote adds an important note to the command's documentation
func (b *CommandBuilder) WithImportantNote(note string) *CommandBuilder {
	b.importantNotes = append(b.importantNotes, note)
	return b
}

// WithValidArgs adds a list of valid arguments that are displayed in completion
func (b *CommandBuilder) WithValidArgs(validArgs []string) *CommandBuilder {
	b.cmd.ValidArgs = validArgs
	return b
}

// WithValidArgsFunction adds a function to validate and generate completion for arguments
func (b *CommandBuilder) WithValidArgsFunction(f func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)) *CommandBuilder {
	b.cmd.ValidArgsFunction = f
	return b
}

// WithFlagCompletionFunc adds completion for a specific flag
func (b *CommandBuilder) WithFlagCompletionFunc(flagName string, f func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)) *CommandBuilder {
	if b.cmd.Flags().Lookup(flagName) != nil {
		_ = b.cmd.RegisterFlagCompletionFunc(flagName, f)
	}
	return b
}

// Build constructs and returns the final command
func (b *CommandBuilder) Build() *cobra.Command {
	// Generate help text if root command is available
	if b.rootCmd != nil {
		fullCommandPath := b.cmd.Use
		if b.cmd.Parent() != nil {
			fullCommandPath = "megaport-cli " + fullCommandPath
		} else {
			fullCommandPath = "megaport-cli " + fullCommandPath
		}

		helpBuilder := &help.CommandHelpBuilder{
			CommandName:    fullCommandPath,
			ShortDesc:      b.cmd.Short,
			LongDesc:       b.cmd.Long,
			RequiredFlags:  b.requiredFlags,
			OptionalFlags:  b.optionalFlags,
			Examples:       b.examples,
			ImportantNotes: b.importantNotes,
			JSONExamples:   b.jsonExamples,
		}
		b.cmd.Long = helpBuilder.Build(b.rootCmd)
	}

	return b.cmd
}
