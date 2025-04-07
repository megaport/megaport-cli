package generate_docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/megaport"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var version string

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

func generateDocs(rootCmd *cobra.Command, args []string) error {
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

	version = megaport.GetGitVersion()
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

	// Use a map to track unique flags by name
	flagMap := make(map[string]FlagInfo)

	// Collect local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		flagMap[flag.Name] = FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: flag.Usage,
			Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
		}
	})

	// For local flags (separate collection)
	var localFlags []FlagInfo
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		localFlags = append(localFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: flag.Usage,
			Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
		})
	})

	// For persistent flags (separate collection)
	var persistentFlags []FlagInfo
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		// Only add to the persistent collection if not already in the flagMap
		if _, exists := flagMap[flag.Name]; !exists {
			flagMap[flag.Name] = FlagInfo{
				Name:        flag.Name,
				Shorthand:   flag.Shorthand,
				Default:     flag.DefValue,
				Description: flag.Usage,
				Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
			}
		}
		persistentFlags = append(persistentFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: flag.Usage,
			Required:    flag.Annotations != nil && flag.Annotations["cobra_annotation_required"] != nil,
		})
	})

	// Convert the map to a slice for the template
	var allFlags []FlagInfo
	for _, flagInfo := range flagMap {
		allFlags = append(allFlags, flagInfo)
	}

	// Sort flags alphabetically for consistent output
	sort.Slice(allFlags, func(i, j int) bool {
		return allFlags[i].Name < allFlags[j].Name
	})

	// Gather subcommands
	var subCommands []string
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden && subCmd.Name() != "help" {
			subCommands = append(subCommands, subCmd.Name())
		}
	}

	baseFileName := filepath.Base(outputPath)
	filePrefix := strings.TrimSuffix(baseFileName, ".md")

	// Determine parent command info
	var parentCommandPath, parentCommandName, parentFilePath string
	hasParent := cmd.Parent() != nil && cmd.Parent().Name() != "megaport-cli"
	if hasParent && cmd.Parent() != nil {
		parentCommandPath = cmd.Parent().CommandPath()
		parentCommandName = cmd.Parent().Name()

		currentPrefix := filePrefix
		lastUnderscoreIndex := strings.LastIndex(currentPrefix, "_")
		if lastUnderscoreIndex >= 0 {
			parentFilePath = currentPrefix[:lastUnderscoreIndex]
		} else {
			parentFilePath = "megaport-cli"
		}

		if parentCommandName == "megaport-cli" {
			parentFilePath = "megaport-cli"
		}
	}

	// Process the long description
	processedLongDesc := output.StripANSIColors(cmd.Long)
	if processedLongDesc != "" {
		lines := strings.Split(processedLongDesc, "\n")
		var formattedLines []string
		inExampleBlock := false
		inExampleSection := false
		inJsonSection := false
		lineAfterExampleHeader := false
		inRequiredFields := false
		inOptionalFields := false
		inImportantNotes := false

		for i, line := range lines {
			trimLine := strings.TrimSpace(line)

			// Convert section headers to level 3 headers
			if trimLine == "Required fields:" {
				inRequiredFields = true
				inOptionalFields = false
				inImportantNotes = false
				// Replace with level 3 header
				formattedLines = append(formattedLines, "### Required Fields")
				continue
			} else if trimLine == "Optional fields:" {
				inRequiredFields = false
				inOptionalFields = true
				inImportantNotes = false
				// Replace with level 3 header
				formattedLines = append(formattedLines, "### Optional Fields")
				continue
			} else if trimLine == "Important notes:" {
				inRequiredFields = false
				inOptionalFields = false
				inImportantNotes = true
				// Replace with level 3 header
				formattedLines = append(formattedLines, "### Important Notes")
				continue
			} else if trimLine == "Example usage:" || trimLine == "Examples:" {
				inRequiredFields = false
				inOptionalFields = false
				inImportantNotes = false
				inExampleSection = true
				lineAfterExampleHeader = true
				// Replace with level 3 header
				formattedLines = append(formattedLines, "### Example Usage")
				continue
			} else if trimLine == "JSON format example:" {
				inRequiredFields = false
				inOptionalFields = false
				inImportantNotes = false
				// Replace with level 3 header
				formattedLines = append(formattedLines, "### JSON Format Example")

				if inExampleBlock {
					formattedLines = append(formattedLines, "```") // Close previous code block
				}

				formattedLines = append(formattedLines, "```json") // Start JSON code block
				inExampleBlock = true
				inJsonSection = true
				inExampleSection = true
				continue
			}

			// Process field entries with bullets and backticks
			if (inRequiredFields || inOptionalFields) && strings.TrimSpace(line) != "" &&
				!strings.HasPrefix(trimLine, "Required fields:") &&
				!strings.HasPrefix(trimLine, "Optional fields:") {

				// Check if this line defines a field
				if strings.Contains(trimLine, ":") {
					parts := strings.SplitN(trimLine, ":", 2)
					if len(parts) == 2 {
						fieldName := strings.TrimSpace(parts[0])
						fieldDesc := strings.TrimSpace(parts[1])

						// Format with bullets, backticks and preserve indentation
						leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
						spacePadding := strings.Repeat(" ", leadingSpaces)
						formattedLines = append(formattedLines,
							fmt.Sprintf("%s- `%s`: %s", spacePadding, fieldName, fieldDesc))
						continue
					}
				}
			}

			// Process important notes with bullets
			if inImportantNotes && strings.TrimSpace(line) != "" &&
				!strings.HasPrefix(trimLine, "Important notes:") {
				// Format with bullets
				leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
				spacePadding := strings.Repeat(" ", leadingSpaces)
				if !strings.HasPrefix(trimLine, "-") {
					formattedLines = append(formattedLines, fmt.Sprintf("%s- %s", spacePadding, trimLine))
				} else {
					formattedLines = append(formattedLines, line)
				}
				continue
			}

			// Start code block after example header
			if lineAfterExampleHeader && trimLine != "" && !strings.HasPrefix(trimLine, "```") {
				formattedLines = append(formattedLines, "```")
				inExampleBlock = true
				lineAfterExampleHeader = false
			}

			// Detect if it's a header line
			isHeaderLine := strings.HasPrefix(trimLine, "#") && len(trimLine) > 1 && trimLine[1] == ' '

			// Close code block if a header is encountered while in example block
			if isHeaderLine && inExampleBlock {
				formattedLines = append(formattedLines, "```")
				inExampleBlock = false
				inExampleSection = false
			}

			// Downgrade headers to level 3 if in an example section or if it contains 'example'
			if isHeaderLine {
				if inExampleSection || strings.Contains(strings.ToLower(trimLine), "example") {
					headerText := strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(trimLine, "#"), " "))
					formattedLines = append(formattedLines, "### "+headerText)
				} else {
					formattedLines = append(formattedLines, line)
				}
				continue
			}

			// Reset example section if a major heading (level 1 or 2) not containing "example" is detected
			if isHeaderLine && !strings.Contains(strings.ToLower(trimLine), "example") &&
				strings.Count(trimLine, "#") <= 2 {
				inExampleSection = false
			}

			// Detect command examples - cover more patterns
			isExampleLine := strings.HasPrefix(trimLine, "megaport-cli") ||
				(cmd.Name() == "buy" && strings.HasPrefix(trimLine, "buy")) ||
				(inExampleSection &&
					trimLine != "" &&
					!isHeaderLine &&
					(strings.Contains(trimLine, "--") ||
						strings.HasPrefix(trimLine, cmd.Name())))

			// Start code block if an example line is detected and not already in one
			if isExampleLine && !inExampleBlock {
				formattedLines = append(formattedLines, "```")
				inExampleBlock = true
			}

			// If we're in an empty line after examples section and code block is open, consider closing it
			if inExampleBlock && trimLine == "" && inExampleSection {
				// Check if this is followed by a non-example line
				hasMoreExamples := false
				for j := i + 1; j < len(lines); j++ {
					nextLine := strings.TrimSpace(lines[j])
					if strings.HasPrefix(nextLine, "megaport-cli") ||
						(cmd.Name() == "buy" && strings.HasPrefix(nextLine, "buy")) ||
						strings.Contains(nextLine, "--") {
						hasMoreExamples = true
						break
					}
				}

				if !hasMoreExamples && !inJsonSection {
					formattedLines = append(formattedLines, "```")
					inExampleBlock = false
					continue
				}
			}

			// Add the line to the output (if we didn't already handle it)
			formattedLines = append(formattedLines, line)
		}

		// Close any open code block at the end
		if inExampleBlock {
			formattedLines = append(formattedLines, "```")
		}

		processedLongDesc = strings.Join(formattedLines, "\n")
	}

	// Detect an Example section in cmd.Long if present separately
	example := cmd.Example
	if example == "" && strings.Contains(cmd.Long, "Example:") {
		parts := strings.Split(cmd.Long, "Example:")
		if len(parts) > 1 {
			example = "Example:\n" + strings.TrimSpace(parts[1])
		}
	}

	data := CommandData{
		Name:               cmd.Name(),
		Description:        output.StripANSIColors(cmd.Short),
		LongDescription:    processedLongDesc,
		Usage:              cmd.UseLine(),
		Example:            output.StripANSIColors(example),
		HasParent:          hasParent,
		ParentCommandPath:  parentCommandPath,
		ParentCommandName:  parentCommandName,
		ParentFilePath:     parentFilePath,
		Aliases:            cmd.Aliases,
		Flags:              allFlags, // Use our deduplicated flags
		LocalFlags:         localFlags,
		PersistentFlags:    persistentFlags,
		HasSubCommands:     len(subCommands) > 0,
		SubCommands:        subCommands,
		IsAvailableCommand: cmd.IsAvailableCommand(),
		FilepathPrefix:     filePrefix,
	}

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

## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
{{ range .Flags }}| ` + "`" + `--{{ .Name }}` + "`" + ` | {{ if .Shorthand }}` + "`" + `-{{ .Shorthand }}` + "`" + `{{ end }} | {{ if .Default }}` + "`" + `{{ .Default }}` + "`" + `{{ end }} | {{ .Description }} | {{ .Required }} |
{{ end }}

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

func AddCommandsTo(rootCmd *cobra.Command) {
	// Set up help builder for the generate-docs command
	genDocsHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli generate-docs",
		ShortDesc:   "Generate markdown documentation for the CLI",
		LongDesc:    "Generate comprehensive markdown documentation for the Megaport CLI.\n\nThis command will extract all command metadata, examples, and annotations to create a set of markdown files that document the entire CLI interface.\n\nThe documentation is organized by command hierarchy, with each command generating its own markdown file containing:\n- Command description\n- Usage examples\n- Available flags\n- Subcommands list\n- Input/output formats (where applicable)",
		Examples: []string{
			"generate-docs ./docs",
		},
		ImportantNotes: []string{
			"The output directory will be created if it doesn't exist",
			"Existing files in the output directory may be overwritten",
			"Hidden commands and 'help' commands are excluded from the documentation",
		},
	}

	// Define genDocsCmd here
	var genDocsCmd = &cobra.Command{
		Use:   "generate-docs [directory]",
		Short: "Generate markdown documentation for the CLI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Pass the root command to generateDocs explicitly
			return generateDocs(rootCmd, args)
		},
	}

	genDocsCmd.Long = genDocsHelp.Build(rootCmd)
	rootCmd.AddCommand(genDocsCmd)
}
