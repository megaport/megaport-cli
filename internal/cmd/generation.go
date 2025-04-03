package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CommandInfo struct {
	Name        string
	Description string
	Path        string
}

type IndexData struct {
	Commands    []CommandInfo
	GeneratedAt string
	Version     string
}

type FlagInfo struct {
	Name        string
	Shorthand   string
	Default     string
	Description string
	Required    bool
}

type CommandData struct {
	Name               string
	Description        string
	LongDescription    string
	Usage              string
	Example            string
	HasParent          bool
	ParentCommandPath  string
	ParentCommandName  string
	ParentFilePath     string
	Aliases            []string
	Flags              []FlagInfo
	LocalFlags         []FlagInfo
	PersistentFlags    []FlagInfo
	HasSubCommands     bool
	SubCommands        []string
	IsAvailableCommand bool
	FilepathPrefix     string // Add this new field
}

// genDocsCmd generates markdown documentation for all commands
var genDocsCmd = &cobra.Command{
	Use:   "generate-docs [directory]",
	Short: "Generate markdown documentation for the CLI",
	Long: `Generate comprehensive markdown documentation for the Megaport CLI.

This command will extract all command metadata, examples, and annotations
to create a set of markdown files that document the entire CLI interface.

The documentation is organized by command hierarchy, with each command
generating its own markdown file containing:
- Command description
- Usage examples
- Available flags
- Subcommands list
- Input/output formats (where applicable)

Example usage:
  megaport-cli generate-docs ./docs
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(generateDocs),
}

func generateDocs(cmd *cobra.Command, args []string) error {
	outputDir := args[0]

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate index.md - a table of contents for all commands
	if err := generateIndex(rootCmd, outputDir); err != nil {
		return fmt.Errorf("failed to generate index: %v", err)
	}

	// Recursively generate docs for all commands
	if err := generateCommandDocs(rootCmd, outputDir, ""); err != nil {
		return fmt.Errorf("failed to generate command docs: %v", err)
	}

	fmt.Printf("Documentation generated in %s\n", outputDir)
	return nil
}

func generateIndex(root *cobra.Command, outputDir string) error {
	// Collect command information
	var commands []CommandInfo
	collectCommands(root, "", &commands)

	data := IndexData{
		Commands:    commands,
		GeneratedAt: time.Now().Format("January 2, 2006"),
		Version:     version,
	}

	// Create index file
	f, err := os.Create(filepath.Join(outputDir, "index.md"))
	if err != nil {
		return err
	}
	defer f.Close()

	// Define the template
	indexTemplate := `# Megaport CLI Documentation

> Generated on {{ .GeneratedAt }} for version {{ .Version }}

## Available Commands

| Command | Description |
|---------|-------------|
{{- range .Commands }}
| [{{ .Name }}]({{ .Path }}) | {{ .Description }} |
{{- end }}
`

	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

func collectCommands(cmd *cobra.Command, parentPath string, commands *[]CommandInfo) {
	if cmd.Hidden || cmd.Name() == "help" {
		return // Skip only hidden commands and help
	}

	cmdPath := cmd.Name()
	if parentPath != "" {
		cmdPath = parentPath + "_" + cmdPath
	}

	*commands = append(*commands, CommandInfo{
		Name:        cmd.CommandPath(),
		Description: cmd.Short,
		Path:        cmdPath + ".md",
	})

	for _, subCmd := range cmd.Commands() {
		collectCommands(subCmd, cmdPath, commands)
	}
}

func generateCommandDocs(cmd *cobra.Command, outputDir, parentPath string) error {
	if cmd.Hidden || cmd.Name() == "help" { // Removed "|| cmd.Name() == "completion"
		return nil // Skip only hidden commands and help
	}

	cmdPath := cmd.Name()
	if parentPath != "" {
		cmdPath = parentPath + "_" + cmdPath
	}

	// Create command doc
	err := generateCommandDoc(cmd, filepath.Join(outputDir, cmdPath+".md"))
	if err != nil {
		return err
	}

	// Generate docs for subcommands
	for _, subCmd := range cmd.Commands() {
		if err := generateCommandDocs(subCmd, outputDir, cmdPath); err != nil {
			return err
		}
	}

	return nil
}

func generateCommandDoc(cmd *cobra.Command, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Collect flags
	var localFlags []FlagInfo
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		localFlags = append(localFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: flag.Usage,
			Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
		})
	})

	var persistentFlags []FlagInfo
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) { // Change to *pflag.Flag
		persistentFlags = append(persistentFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: flag.Usage,
			Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
		})
	})

	// Combine all flags
	var allFlags []FlagInfo
	allFlags = append(allFlags, localFlags...)
	allFlags = append(allFlags, persistentFlags...)

	// Get subcommands
	var subCommands []string
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden && subCmd.Name() != "help" {
			subCommands = append(subCommands, subCmd.Name())
		}
	}

	baseFileName := filepath.Base(outputPath)
	filePrefix := strings.TrimSuffix(baseFileName, ".md")

	// Now we can use filePrefix to determine parent file path
	var parentCommandPath, parentCommandName, parentFilePath string
	hasParent := cmd.Parent() != nil && cmd.Parent().Name() != "megaport-cli"
	if hasParent && cmd.Parent() != nil {
		parentCommandPath = cmd.Parent().CommandPath()
		parentCommandName = cmd.Parent().Name()

		// Build the parent file path using the same pattern as we use for commands
		// Extract the current file prefix up to the last underscore
		currentPrefix := filePrefix
		lastUnderscoreIndex := strings.LastIndex(currentPrefix, "_")
		if lastUnderscoreIndex >= 0 {
			// For nested commands, remove the last segment
			parentFilePath = currentPrefix[:lastUnderscoreIndex]
		} else {
			// For top level commands, the parent is megaport-cli
			parentFilePath = "megaport-cli"
		}

		// If the parent is the root command (megaport-cli), handle it specially
		if parentCommandName == "megaport-cli" {
			parentFilePath = "megaport-cli"
		}
	}

	// Process the long description to format examples as code blocks
	// Process the long description to format examples as code blocks
	processedLongDesc := cmd.Long
	if processedLongDesc != "" {
		lines := strings.Split(processedLongDesc, "\n")
		var formattedLines []string
		inExampleBlock := false
		inExampleSection := false // Track if we're in an examples section

		for i, line := range lines {
			trimLine := strings.TrimSpace(line)

			// Format field descriptions with backticks
			if strings.HasPrefix(trimLine, "-") && strings.Contains(trimLine, ":") {
				// Extract the field name and description
				dashParts := strings.SplitN(trimLine, ":", 2)
				if len(dashParts) == 2 {
					// Get the field name (after the dash, before the colon)
					fieldPart := strings.TrimSpace(dashParts[0])
					fieldName := strings.TrimSpace(strings.TrimPrefix(fieldPart, "-"))

					// Get the description (after the colon)
					description := strings.TrimSpace(dashParts[1])

					// Check if the field includes a requirement note in parentheses
					requiredNote := ""
					if strings.Contains(fieldName, "(") && strings.Contains(fieldName, ")") {
						// Extract the requirement note
						nameAndReq := strings.SplitN(fieldName, "(", 2)
						if len(nameAndReq) == 2 {
							fieldName = strings.TrimSpace(nameAndReq[0])
							requiredNote = " (" + strings.TrimSuffix(nameAndReq[1], ")") + ")"
						}
					}

					// Format with backticks around the field name
					formattedLine := "- `" + fieldName + "`" + requiredNote + ": " + description
					formattedLines = append(formattedLines, formattedLine)
					continue
				}
			}

			// Detect if we've entered an examples section
			if strings.Contains(strings.ToLower(trimLine), "example") &&
				(strings.HasSuffix(trimLine, ":") || strings.HasSuffix(trimLine, "usage") ||
					strings.HasPrefix(trimLine, "#")) {
				inExampleSection = true
			}

			// Detect if this is a header line (starts with # and has text after it)
			isHeaderLine := strings.HasPrefix(trimLine, "#") && len(trimLine) > 1 && trimLine[1] == ' '

			// If we hit a header and we're in a code block, close the code block first
			if isHeaderLine && inExampleBlock {
				formattedLines = append(formattedLines, "```")
				inExampleBlock = false
			}

			// Handle headers based on whether they are in an example section
			if isHeaderLine {
				// If we're in an example section OR the header contains "example", make it level 3
				if inExampleSection || strings.Contains(strings.ToLower(trimLine), "example") {
					// Convert to level 3 header (###) regardless of original level
					headerText := strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(trimLine, "#"), " "))
					formattedLines = append(formattedLines, "### "+headerText)
				} else {
					// Keep other headers as they are
					formattedLines = append(formattedLines, line)
				}
				continue
			}

			// Reset example section status if we hit a major heading that's not an example
			if isHeaderLine && !strings.Contains(strings.ToLower(trimLine), "example") &&
				strings.Count(trimLine, "#") <= 2 { // Level 1 or 2 heading
				inExampleSection = false
			}

			// Detect example command lines by common patterns
			isExampleLine := strings.HasPrefix(trimLine, "megaport-cli") ||
				(i > 0 && strings.Contains(strings.ToLower(lines[i-1]), "example") &&
					!strings.HasPrefix(trimLine, "#") && trimLine != "")

			// Start a code block before an example if not already in one
			if isExampleLine && !inExampleBlock {
				formattedLines = append(formattedLines, "```")
				inExampleBlock = true
			}

			// End a code block after an example if there's a blank line or end of section
			if inExampleBlock && (trimLine == "" ||
				(i < len(lines)-1 && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "#"))) {
				formattedLines = append(formattedLines, line)
				formattedLines = append(formattedLines, "```")
				if trimLine == "" {
					formattedLines = append(formattedLines, "") // Keep the blank line
				}
				inExampleBlock = false
				continue
			}

			// Preserve special case for sections like "Examples:"
			if strings.HasPrefix(trimLine, "Example") && strings.HasSuffix(trimLine, ":") {
				if inExampleBlock {
					formattedLines = append(formattedLines, "```")
					inExampleBlock = false
				}
				formattedLines = append(formattedLines, line)
				continue
			}

			// Add the line if not already handled above
			formattedLines = append(formattedLines, line)
		}

		// Close any open code block at the end
		if inExampleBlock {
			formattedLines = append(formattedLines, "```")
		}

		processedLongDesc = strings.Join(formattedLines, "\n")
	}

	// Also handle the specific Example section if present
	example := cmd.Example
	if example == "" && strings.Contains(cmd.Long, "Example:") {
		parts := strings.Split(cmd.Long, "Example:")
		if len(parts) > 1 {
			example = "Example:\n" + strings.TrimSpace(parts[1])
		}
	}

	data := CommandData{
		Name:               cmd.Name(),
		Description:        cmd.Short,
		LongDescription:    processedLongDesc,
		Usage:              cmd.UseLine(),
		Example:            example,
		HasParent:          hasParent,
		ParentCommandPath:  parentCommandPath,
		ParentCommandName:  parentCommandName,
		ParentFilePath:     parentFilePath,
		Aliases:            cmd.Aliases,
		Flags:              allFlags,
		LocalFlags:         localFlags,
		PersistentFlags:    persistentFlags,
		HasSubCommands:     len(subCommands) > 0,
		SubCommands:        subCommands,
		IsAvailableCommand: cmd.IsAvailableCommand(),
		FilepathPrefix:     filePrefix,
	}
	// Create the command doc file
	cmdTemplate := `# {{ .Name }}

{{ .Description }}

{{ if .LongDescription }}## Description

{{ .LongDescription }}
{{ end }}

## Usage

` + "```" + `
{{ .Usage }}
` + "```" + `

{{ if .Example }}## Examples

` + "```" + `
{{ .Example }}
` + "```" + `{{ end }}

{{ if .HasParent }}## Parent Command

* [{{ .ParentCommandPath }}]({{ .ParentFilePath }}.md)
{{ end }}

{{ if .Aliases }}## Aliases

{{ range .Aliases }}* {{ . }}
{{ end }}{{ end }}

{{ if .Flags }}## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
{{ range .Flags }}| ` + "`" + `--{{ .Name }}` + "`" + ` | {{ if .Shorthand }}` + "`" + `-{{ .Shorthand }}` + "`" + `{{ end }} | {{ if .Default }}` + "`" + `{{ .Default }}` + "`" + `{{ end }} | {{ .Description }} | {{ .Required }} |
{{ end }}{{ end }}

{{ if .HasSubCommands }}## Subcommands

{{ range .SubCommands }}* [{{ . }}]({{ $.FilepathPrefix }}_{{ . }}.md)
{{ end }}{{ end }}
`

	tmpl, err := template.New("command").Parse(cmdTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

func init() {
	rootCmd.AddCommand(genDocsCmd)
}
