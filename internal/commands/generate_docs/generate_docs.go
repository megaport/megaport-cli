package generate_docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/version"
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
	FilepathPrefix     string
}

func generateDocs(rootCmd *cobra.Command, args []string) error {
	outputDir := args[0]

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	spinner := output.NewSpinner(false)
	spinner.Start("Generating CLI documentation...")

	if err := generateIndex(rootCmd, outputDir); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate index: %v", err)
	}

	if err := generateCommandDocs(rootCmd, outputDir, ""); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate command docs: %v", err)
	}

	spinner.StopWithSuccess(fmt.Sprintf("Documentation successfully generated in %s", outputDir))
	return nil
}

func generateIndex(root *cobra.Command, outputDir string) error {
	var commands []CommandInfo
	collectCommands(root, "", &commands)

	v := version.GetGitVersion()
	data := IndexData{
		Commands:    commands,
		GeneratedAt: time.Now().Format("January 2, 2006"),
		Version:     v,
	}

	f, err := os.Create(filepath.Join(outputDir, "index.md"))
	if err != nil {
		return err
	}
	defer f.Close()

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
		return
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
	if cmd.Hidden || cmd.Name() == "help" {
		return nil
	}

	cmdPath := cmd.Name()
	if parentPath != "" {
		cmdPath = parentPath + "_" + cmdPath
	}

	err := generateCommandDoc(cmd, filepath.Join(outputDir, cmdPath+".md"))
	if err != nil {
		return err
	}

	for _, subCmd := range cmd.Commands() {
		if err := generateCommandDocs(subCmd, outputDir, cmdPath); err != nil {
			return err
		}
	}

	return nil
}

func collectFlags(cmd *cobra.Command) ([]FlagInfo, []FlagInfo, []FlagInfo) {
	flagMap := make(map[string]FlagInfo)

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		cmdFlag := cmd.Flags().Lookup(flag.Name)
		required := false
		if cmdFlag != nil && cmdFlag.Annotations != nil {
			_, required = cmdFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
		}
		if !required && strings.Contains(flag.Usage, "[required]") {
			required = true
		}
		if !required && strings.Contains(strings.ToLower(flag.Usage), "required flag") {
			required = true
		}
		flagMap[flag.Name] = FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: strings.TrimSpace(strings.ReplaceAll(flag.Usage, "[required]", "")),
			Required:    required,
		}
	})

	var localFlags []FlagInfo
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		cmdFlag := cmd.Flags().Lookup(flag.Name)
		required := false
		if cmdFlag != nil && cmdFlag.Annotations != nil {
			_, required = cmdFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
		}
		if !required && strings.Contains(flag.Usage, "[required]") {
			required = true
		}
		if !required && strings.Contains(strings.ToLower(flag.Usage), "required flag") {
			required = true
		}
		localFlags = append(localFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: strings.TrimSpace(strings.ReplaceAll(flag.Usage, "[required]", "")),
			Required:    required,
		})
	})

	var persistentFlags []FlagInfo
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		cmdFlag := cmd.PersistentFlags().Lookup(flag.Name)
		required := false
		if cmdFlag != nil && cmdFlag.Annotations != nil {
			_, required = cmdFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
		}
		if !required && strings.Contains(flag.Usage, "[required]") {
			required = true
		}
		if !required && strings.Contains(strings.ToLower(flag.Usage), "required flag") {
			required = true
		}
		if _, exists := flagMap[flag.Name]; !exists {
			flagMap[flag.Name] = FlagInfo{
				Name:        flag.Name,
				Shorthand:   flag.Shorthand,
				Default:     flag.DefValue,
				Description: strings.TrimSpace(strings.ReplaceAll(flag.Usage, "[required]", "")),
				Required:    required,
			}
		}
		persistentFlags = append(persistentFlags, FlagInfo{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Default:     flag.DefValue,
			Description: strings.TrimSpace(strings.ReplaceAll(flag.Usage, "[required]", "")),
			Required:    required,
		})
	})

	var allFlags []FlagInfo
	for _, flagInfo := range flagMap {
		allFlags = append(allFlags, flagInfo)
	}

	sort.Slice(allFlags, func(i, j int) bool {
		return allFlags[i].Name < allFlags[j].Name
	})

	return allFlags, localFlags, persistentFlags
}

func gatherSubcommands(cmd *cobra.Command) []string {
	var subCommands []string
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden && subCmd.Name() != "help" {
			subCommands = append(subCommands, subCmd.Name())
		}
	}
	return subCommands
}

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

func formatFieldLine(line string) string {
	trimLine := strings.TrimSpace(line)

	if strings.Contains(trimLine, ":") {
		parts := strings.SplitN(trimLine, ":", 2)
		if len(parts) == 2 {
			fieldName := strings.TrimSpace(parts[0])
			fieldDesc := strings.TrimSpace(parts[1])
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
			spacePadding := strings.Repeat(" ", leadingSpaces)
			return fmt.Sprintf("%s- `%s`: %s", spacePadding, fieldName, fieldDesc)
		}
	}
	return line
}

func formatNoteLine(line string) string {
	trimLine := strings.TrimSpace(line)

	if trimLine != "" && !strings.HasPrefix(trimLine, "-") {
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		spacePadding := strings.Repeat(" ", leadingSpaces)
		return fmt.Sprintf("%s- %s", spacePadding, trimLine)
	}
	return line
}

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

		if (inRequiredFields || inOptionalFields) && trimLine != "" &&
			!strings.HasPrefix(trimLine, "Required fields:") &&
			!strings.HasPrefix(trimLine, "Optional fields:") {
			formattedLines = append(formattedLines, formatFieldLine(line))
			continue
		}

		if inImportantNotes && trimLine != "" &&
			!strings.HasPrefix(trimLine, "Important notes:") {
			formattedLines = append(formattedLines, formatNoteLine(line))
			continue
		}

		if lineAfterExampleHeader && trimLine != "" && !strings.HasPrefix(trimLine, "```") {
			formattedLines = append(formattedLines, "```sh")
			inExampleBlock = true
			lineAfterExampleHeader = false
		}

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

		if isHeaderLine && !strings.Contains(strings.ToLower(trimLine), "example") &&
			strings.Count(trimLine, "#") <= 2 {
			inExampleSection = false
		}

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

	if inExampleBlock {
		formattedLines = append(formattedLines, "```")
	}

	return strings.Join(formattedLines, "\n")
}

func generateCommandDoc(cmd *cobra.Command, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	allFlags, localFlags, persistentFlags := collectFlags(cmd)
	subCommands := gatherSubcommands(cmd)
	baseFileName := filepath.Base(outputPath)
	filePrefix := strings.TrimSuffix(baseFileName, ".md")
	hasParent, parentCommandPath, parentCommandName, parentFilePath :=
		determineParentInfo(cmd, filePrefix)

	processedLongDesc := output.StripANSIColors(cmd.Long)
	processedLongDesc = processDescription(processedLongDesc, cmd.Name())

	example := cmd.Example
	if example == "" && strings.Contains(cmd.Long, "Example:") {
		parts := strings.Split(cmd.Long, "Example:")
		if len(parts) > 1 {
			example = "Example:\n" + strings.TrimSpace(parts[1])
		}
	}
	example = output.StripANSIColors(example)

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

	cmdTemplate := getCommandTemplate()
	tmpl, err := template.New("command").Parse(cmdTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

func getCommandTemplate() string {
	return `# {{ .Name }}

{{ .Description }}

{{ if .LongDescription }}## Description

{{ .LongDescription }}{{ end }}

## Usage

` + "```sh" + `
{{ .Usage }}
` + "```" + `

{{ if .Example }}## Examples

` + "```sh" + `
{{ .Example }}
` + "```" + `{{ end }}{{ if .HasParent }}
## Parent Command

* [{{ .ParentCommandPath }}]({{ .ParentFilePath }}.md){{ end }}
{{ if .Aliases }}
## Aliases

{{ range .Aliases }}* {{ . }}
{{ end }}{{ end }}## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
{{ range .Flags }}| ` + "`" + `--{{ .Name }}` + "`" + ` | {{ if .Shorthand }}` + "`" + `-{{ .Shorthand }}` + "`" + `{{ end }} | {{ if .Default }}` + "`" + `{{ .Default }}` + "`" + `{{ end }} | {{ .Description }} | {{ .Required }} |
{{ end }}{{ if .HasSubCommands }}
## Subcommands
{{ range .SubCommands }}* [{{ . }}]({{ $.FilepathPrefix }}_{{ . }}.md)
{{ end }}{{ end }}
`
}

func AddCommandsTo(rootCmd *cobra.Command) {
	genDocsCmd := cmdbuilder.NewCommand("generate-docs", "Generate markdown documentation for the CLI").
		WithArgs(cobra.ExactArgs(1)).
		WithRunFunc(func(cmd *cobra.Command, args []string) error {
			return generateDocs(rootCmd, args)
		}).
		WithExample("megaport-cli generate-docs ./docs").
		WithImportantNote("The output directory will be created if it doesn't exist").
		WithImportantNote("Existing files in the output directory may be overwritten").
		WithImportantNote("Hidden commands and 'help' commands are excluded from the documentation").
		WithLongDesc(
			"Generate comprehensive markdown documentation for the Megaport CLI.\n\n" +
				"This command will extract all command metadata, examples, and annotations to " +
				"create a set of markdown files that document the entire CLI interface.\n\n" +
				"The documentation is organized by command hierarchy, with each command generating " +
				"its own markdown file containing:\n" +
				"- Command description\n" +
				"- Usage examples\n" +
				"- Available flags\n" +
				"- Subcommands list\n" +
				"- Input/output formats (where applicable)").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(genDocsCmd)
}
