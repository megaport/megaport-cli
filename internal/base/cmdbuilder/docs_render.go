package cmdbuilder

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

// DocsDirectory is the fallback location for markdown documentation files
var DocsDirectory = "./docs"

// embeddedDocsFS holds the embedded documentation files
var embeddedDocsFS embed.FS

// RegisterEmbeddedDocs registers the embedded documentation filesystem
func RegisterEmbeddedDocs(docs embed.FS) {
	embeddedDocsFS = docs
}

// FindDocContent returns the raw markdown documentation for a command,
// preferring the embedded docs and falling back to the local docs directory.
func FindDocContent(cmd *cobra.Command) ([]byte, error) {
	cmdPath := getCommandPath(cmd)
	docName := cmdPath + ".md"

	// First try to read from embedded docs
	embeddedPath := filepath.Join("docs", docName)
	if content, err := embeddedDocsFS.ReadFile(embeddedPath); err == nil {
		return content, nil
	}

	// If embedded file not found, try local docs directory as fallback
	docPath := filepath.Join(DocsDirectory, docName)
	content, err := os.ReadFile(docPath)
	if err != nil {
		return nil, fmt.Errorf("documentation file not found for %s: %w", cmdPath, err)
	}

	return content, nil
}

// RenderMarkdown renders markdown content using Glamour
func RenderMarkdown(content []byte) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	rendered, err := renderer.Render(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to render documentation: %w", err)
	}

	return rendered, nil
}

// ShowDocumentation displays rendered documentation for a command
func ShowDocumentation(cmd *cobra.Command) error {
	content, err := FindDocContent(cmd)
	if err != nil {
		return err
	}

	rendered, err := RenderMarkdown(content)
	if err != nil {
		return err
	}

	// Print the rendered documentation
	fmt.Println(rendered)
	return nil
}

// getCommandPath returns the file path-style name for a command (e.g., "megaport-cli_mcr_buy")
func getCommandPath(cmd *cobra.Command) string {
	if cmd.Parent() == nil {
		return "megaport-cli"
	}

	// Build the full path
	path := cmd.Name()
	parent := cmd.Parent()

	for parent != nil && parent.Name() != "" && parent.Name() != "megaport-cli" {
		path = parent.Name() + "_" + path
		parent = parent.Parent()
	}

	return "megaport-cli_" + path
}
