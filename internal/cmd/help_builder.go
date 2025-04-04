package cmd

import (
	"fmt"
	"strings"
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
}

// Build constructs the full formatted help text
func (b *CommandHelpBuilder) Build() string {
	var sb strings.Builder

	// Command description
	sb.WriteString(b.LongDesc)
	sb.WriteString("\n\n")

	// Input methods section (if any flags defined)
	if len(b.RequiredFlags) > 0 || len(b.OptionalFlags) > 0 {
		sb.WriteString("You can provide details in one of three ways:\n\n")
		sb.WriteString("1. Interactive Mode (with --interactive):\n")
		sb.WriteString("   The command will prompt you for each required and optional field.\n\n")
		sb.WriteString("2. Flag Mode:\n")
		sb.WriteString("   Provide all required fields as flags.\n\n")
		sb.WriteString("3. JSON Mode:\n")
		sb.WriteString("   Provide a JSON string or file with all required fields:\n")
		sb.WriteString("   --json <json-string> or --json-file <path>\n\n")
	}

	// Required flags section
	if len(b.RequiredFlags) > 0 {
		sb.WriteString("Required fields:\n")
		for flag, desc := range b.RequiredFlags {
			sb.WriteString("  ")
			sb.WriteString(FormatRequiredFlag(flag, desc))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Optional flags section
	if len(b.OptionalFlags) > 0 {
		sb.WriteString("Optional fields:\n")
		for flag, desc := range b.OptionalFlags {
			sb.WriteString("  ")
			sb.WriteString(FormatOptionalFlag(flag, desc))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Important notes section
	if len(b.ImportantNotes) > 0 {
		sb.WriteString("Important notes:\n")
		for _, note := range b.ImportantNotes {
			sb.WriteString("- ")
			sb.WriteString(note)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Examples section
	if len(b.Examples) > 0 {
		sb.WriteString("Example usage:\n\n")
		for _, example := range b.Examples {
			sb.WriteString("  ")
			sb.WriteString(FormatExample(fmt.Sprintf("%s %s", b.CommandName, example)))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// JSON examples section
	if len(b.JSONExamples) > 0 {
		sb.WriteString("JSON format example:\n")
		for _, example := range b.JSONExamples {
			sb.WriteString(FormatJSONExample(example))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
