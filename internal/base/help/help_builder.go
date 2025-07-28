package help

import (
	"os"
	"sort"
	"strings"

	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

// CommandHelpBuilder helps construct formatted help text for commands
type CommandHelpBuilder struct {
	CommandName    string
	ShortDesc      string
	LongDesc       string
	Examples       []string
	RequiredFlags  map[string]string // flag name -> description
	OptionalFlags  map[string]string // flag name -> description
	JSONExamples   []string
	ImportantNotes []string
	DisableColor   bool // Whether to disable colored output
}

// Build constructs the full formatted help text
func (b *CommandHelpBuilder) Build(rootCmd *cobra.Command) string {
	var sb strings.Builder
	hasSections := false

	// Better flag detection - check both parsed flags and raw args
	// Better flag detection - check both parsed flags and raw args
	noColor := b.DisableColor

	// Then check environment variables
	if !noColor {
		if _, exists := os.LookupEnv("NO_COLOR"); exists {
			noColor = true
		}
	}

	// Check command line args directly - more reliable with --help
	if !noColor && os.Args != nil {
		for _, arg := range os.Args {
			if arg == "--no-color" || arg == "-no-color" {
				noColor = true
				break
			}
		}
	}

	// Finally try the flag if available
	if !noColor && rootCmd != nil {
		flagVal, err := rootCmd.PersistentFlags().GetBool("no-color")
		if err == nil {
			noColor = flagVal
		}
	}

	// Command description
	if b.LongDesc != "" {
		if noColor {
			sb.WriteString(b.LongDesc)
		} else {
			sb.WriteString(ansi.Color(b.LongDesc, "green+b"))
		}
		sb.WriteString("\n\n")
		hasSections = true
	}

	// Required flags section
	if len(b.RequiredFlags) > 0 {
		if noColor {
			sb.WriteString("Required fields:\n")
		} else {
			sb.WriteString(ansi.Color("Required fields:\n", "yellow+b"))
		}

		// Sort the flag names for consistent output
		var flagNames []string
		for flag := range b.RequiredFlags {
			flagNames = append(flagNames, flag)
		}
		sort.Strings(flagNames)

		// Display flags in sorted order
		for _, flag := range flagNames {
			desc := b.RequiredFlags[flag]
			sb.WriteString("  ")
			sb.WriteString(flag)
			sb.WriteString(": ")
			sb.WriteString(desc)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		hasSections = true
	}

	// Optional flags section
	if len(b.OptionalFlags) > 0 {
		if noColor {
			sb.WriteString("Optional fields:\n")
		} else {
			sb.WriteString(ansi.Color("Optional fields:\n", "yellow+b"))
		}

		// Sort the flag names for consistent output
		var flagNames []string
		for flag := range b.OptionalFlags {
			flagNames = append(flagNames, flag)
		}
		sort.Strings(flagNames)

		// Display flags in sorted order
		for _, flag := range flagNames {
			desc := b.OptionalFlags[flag]
			sb.WriteString("  ")
			sb.WriteString(flag)
			sb.WriteString(": ")
			sb.WriteString(desc)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		hasSections = true
	}

	// Important notes section
	if len(b.ImportantNotes) > 0 {
		if noColor {
			sb.WriteString("Important notes:\n")
		} else {
			sb.WriteString(ansi.Color("Important notes:\n", "yellow+b"))
		}
		for _, note := range b.ImportantNotes {
			sb.WriteString("  - ")
			sb.WriteString(note)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		hasSections = true
	}

	// Examples section
	if len(b.Examples) > 0 {
		if noColor {
			sb.WriteString("Example usage:\n\n")
		} else {
			sb.WriteString(ansi.Color("Example usage:\n\n", "cyan+b"))
		}
		for _, example := range b.Examples {
			sb.WriteString("  ")
			sb.WriteString(example)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		hasSections = true
	}

	// JSON Examples section
	if len(b.JSONExamples) > 0 {
		if noColor {
			sb.WriteString("JSON format example:\n")
		} else {
			sb.WriteString(ansi.Color("JSON format example:\n", "cyan+b"))
		}
		for _, example := range b.JSONExamples {
			if noColor {
				sb.WriteString(example)
			} else {
				sb.WriteString(ansi.Color(example, "green"))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		hasSections = true
	}

	// Return the final help text, ensuring we don't end with excessive newlines
	result := sb.String()

	// Remove trailing newlines (but keep one if there was at least one section)
	result = strings.TrimRight(result, "\n")
	if hasSections {
		result += "\n"
	} else {
		result = "\n"
	}

	return result
}
