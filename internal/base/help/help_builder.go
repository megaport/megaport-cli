package help

import (
	"os"
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

	// Check for no-color flag in multiple ways if DisableColor isn't explicitly set
	noColor := b.DisableColor

	// If DisableColor wasn't set explicitly, check environment and flags
	if !noColor {
		// Method 1: Check environment variable
		if _, exists := os.LookupEnv("NO_COLOR"); exists {
			noColor = true
		}

		// Method 2: Check command line args for --no-color
		for _, arg := range os.Args {
			if arg == "--no-color" {
				noColor = true
				break
			}
		}

		// Method 3: Try to get from cobra rootCmd if available
		if rootCmd != nil {
			if flagVal, err := rootCmd.PersistentFlags().GetBool("no-color"); err == nil {
				noColor = flagVal
			}
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
		for flag, desc := range b.RequiredFlags {
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
		for flag, desc := range b.OptionalFlags {
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
