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

// collectFlags gathers and deduplicates flags from the command
func collectFlags(cmd *cobra.Command) ([]FlagInfo, []FlagInfo, []FlagInfo) {
	// Use a map to track unique flags by name
	flagMap := make(map[string]FlagInfo)

	// Collect all flags
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

	return allFlags, localFlags, persistentFlags
}

// gatherSubcommands collects visible subcommands
func gatherSubcommands(cmd *cobra.Command) []string {
	var subCommands []string
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden && subCmd.Name() != "help" {
			subCommands = append(subCommands, subCmd.Name())
		}
	}
	return subCommands
}

// determineParentInfo calculates parent command relationships
func determineParentInfo(cmd *cobra.Command, filePrefix string) (bool, string, string, string) {
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

	return hasParent, parentCommandPath, parentCommandName, parentFilePath
}

// formatSection handles section formatting based on section type
func formatSection(section string) string {
	switch section {
	case "Required fields:":
		return "### Required Fields"
	case "Optional fields:":
		return "### Optional Fields"
	case "Important notes:":
		return "### Important Notes"
	case "Example usage:", "Examples:":
		return "### Example Usage"
	case "JSON format example:":
		return "### JSON Format Example"
	default:
		return section
	}
}

// formatFieldLine formats field definitions with bullets and backticks
func formatFieldLine(line string) string {
	trimLine := strings.TrimSpace(line)

	if strings.Contains(trimLine, ":") {
		parts := strings.SplitN(trimLine, ":", 2)
		if len(parts) == 2 {
			fieldName := strings.TrimSpace(parts[0])
			fieldDesc := strings.TrimSpace(parts[1])

			// Format with bullets and backticks
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
			spacePadding := strings.Repeat(" ", leadingSpaces)
			return fmt.Sprintf("%s- `%s`: %s", spacePadding, fieldName, fieldDesc)
		}
	}
	return line
}

// formatNoteLine formats important notes with bullets
func formatNoteLine(line string) string {
	trimLine := strings.TrimSpace(line)

	if trimLine != "" && !strings.HasPrefix(trimLine, "-") {
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		spacePadding := strings.Repeat(" ", leadingSpaces)
		return fmt.Sprintf("%s- %s", spacePadding, trimLine)
	}
	return line
}

// processDescription handles the main formatting of command descriptions
func processDescription(description string, cmdName string) string {
	if description == "" {
		return ""
	}

	lines := strings.Split(description, "\n")
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

		// Check for section headers
		switch trimLine {
		case "Required fields:":
			inRequiredFields = true
			inOptionalFields = false
			inImportantNotes = false
			formattedLines = append(formattedLines, formatSection(trimLine))
			continue
		case "Optional fields:":
			inRequiredFields = false
			inOptionalFields = true
			inImportantNotes = false
			formattedLines = append(formattedLines, formatSection(trimLine))
			continue
		case "Important notes:":
			inRequiredFields = false
			inOptionalFields = false
			inImportantNotes = true
			formattedLines = append(formattedLines, formatSection(trimLine))
			continue
		case "Example usage:", "Examples:":
			inRequiredFields = false
			inOptionalFields = false
			inImportantNotes = false
			inExampleSection = true
			lineAfterExampleHeader = true
			formattedLines = append(formattedLines, formatSection(trimLine))
			continue
		case "JSON format example:":
			inRequiredFields = false
			inOptionalFields = false
			inImportantNotes = false
			formattedLines = append(formattedLines, formatSection(trimLine))

			if inExampleBlock {
				formattedLines = append(formattedLines, "```")
			}

			formattedLines = append(formattedLines, "```json")
			inExampleBlock = true
			inJsonSection = true
			inExampleSection = true
			continue
		}

		// Process field entries
		if (inRequiredFields || inOptionalFields) && trimLine != "" &&
			!strings.HasPrefix(trimLine, "Required fields:") &&
			!strings.HasPrefix(trimLine, "Optional fields:") {
			formattedLines = append(formattedLines, formatFieldLine(line))
			continue
		}

		// Process important notes
		if inImportantNotes && trimLine != "" &&
			!strings.HasPrefix(trimLine, "Important notes:") {
			formattedLines = append(formattedLines, formatNoteLine(line))
			continue
		}

		// Start code block after example header
		if lineAfterExampleHeader && trimLine != "" && !strings.HasPrefix(trimLine, "```") {
			formattedLines = append(formattedLines, "```")
			inExampleBlock = true
			lineAfterExampleHeader = false
		}

		// Handle header lines
		isHeaderLine := strings.HasPrefix(trimLine, "#") && len(trimLine) > 1 && trimLine[1] == ' '

		if isHeaderLine && inExampleBlock {
			formattedLines = append(formattedLines, "```")
			inExampleBlock = false
			inExampleSection = false
		}

		if isHeaderLine {
			if inExampleSection || strings.Contains(strings.ToLower(trimLine), "example") {
				headerText := strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(trimLine, "#"), " "))
				formattedLines = append(formattedLines, "### "+headerText)
			} else {
				formattedLines = append(formattedLines, line)
			}
			continue
		}

		// Reset example section
		if isHeaderLine && !strings.Contains(strings.ToLower(trimLine), "example") &&
			strings.Count(trimLine, "#") <= 2 {
			inExampleSection = false
		}

		// Detect command examples
		isExampleLine := strings.HasPrefix(trimLine, "megaport-cli") ||
			(cmdName == "buy" && strings.HasPrefix(trimLine, "buy")) ||
			(inExampleSection &&
				trimLine != "" &&
				!isHeaderLine &&
				(strings.Contains(trimLine, "--") ||
					strings.HasPrefix(trimLine, cmdName)))

		if isExampleLine && !inExampleBlock {
			formattedLines = append(formattedLines, "```")
			inExampleBlock = true
		}

		// Handle end of examples
		if inExampleBlock && trimLine == "" && inExampleSection {
			hasMoreExamples := false
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				nextLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(nextLine, "megaport-cli") ||
					(cmdName == "buy" && strings.HasPrefix(nextLine, "buy")) ||
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

		formattedLines = append(formattedLines, line)
	}

	// Close any open code block
	if inExampleBlock {
		formattedLines = append(formattedLines, "```")
	}

	return strings.Join(formattedLines, "\n")
}

// generateCommandDoc creates documentation for a single command
func generateCommandDoc(cmd *cobra.Command, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Collect flags
	allFlags, localFlags, persistentFlags := collectFlags(cmd)

	// Gather subcommands
	subCommands := gatherSubcommands(cmd)

	// Calculate file paths
	baseFileName := filepath.Base(outputPath)
	filePrefix := strings.TrimSuffix(baseFileName, ".md")

	// Determine parent command info
	hasParent, parentCommandPath, parentCommandName, parentFilePath :=
		determineParentInfo(cmd, filePrefix)

	// Process the long description
	processedLongDesc := output.StripANSIColors(cmd.Long)
	processedLongDesc = processDescription(processedLongDesc, cmd.Name())

	// Process examples
	example := cmd.Example
	if example == "" && strings.Contains(cmd.Long, "Example:") {
		parts := strings.Split(cmd.Long, "Example:")
		if len(parts) > 1 {
			example = "Example:\n" + strings.TrimSpace(parts[1])
		}
	}
	example = output.StripANSIColors(example)

	// Prepare template data
	data := CommandData{
		Name:               cmd.Name(),
		Description:        output.StripANSIColors(cmd.Short),
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

	// Apply template
	cmdTemplate := getCommandTemplate()
	tmpl, err := template.New("command").Parse(cmdTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

// getCommandTemplate returns the markdown template for command docs
func getCommandTemplate() string {
	return `# {{ .Name }}

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
