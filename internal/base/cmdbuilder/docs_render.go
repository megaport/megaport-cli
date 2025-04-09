package cmdbuilder

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// FindDocFile locates the markdown file for a specific command
func FindDocFile(cmd *cobra.Command) (string, error) {
	cmdPath := getCommandPath(cmd)
	docName := cmdPath + ".md"

	// First try to read from embedded docs
	embeddedPath := filepath.Join("docs", docName)
	content, err := embeddedDocsFS.ReadFile(embeddedPath)
	if err == nil {
		// Create a temporary file to store the content for rendering
		tempFile, err := os.CreateTemp("", "megaport-docs-*.md")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer tempFile.Close()

		if _, err := tempFile.Write(content); err != nil {
			return "", fmt.Errorf("failed to write to temporary file: %w", err)
		}

		return tempFile.Name(), nil
	}

	// If embedded file not found, try local docs directory as fallback
	docPath := filepath.Join(DocsDirectory, docName)

	// Check if the file exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return "", fmt.Errorf("documentation file not found for %s: %w", cmdPath, err)
	}

	return docPath, nil
}

// RenderDocFile reads and renders a markdown file using Glamour
func RenderDocFile(filePath string) (string, error) {
	// Read the markdown file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read documentation file: %w", err)
	}

	// Create a glamour renderer with default style
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	// Render the markdown content
	rendered, err := renderer.Render(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to render documentation: %w", err)
	}

	return rendered, nil
}

// ShowDocumentation displays rendered documentation for a command
func ShowDocumentation(cmd *cobra.Command) error {
	docPath, err := FindDocFile(cmd)
	if err != nil {
		return err
	}

	// If we created a temporary file, ensure it gets deleted
	if strings.Contains(docPath, "megaport-docs-") {
		defer os.Remove(docPath)
	}

	rendered, err := RenderDocFile(docPath)
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
