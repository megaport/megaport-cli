//go:build !js || !wasm

package cmdbuilder

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

// DocsDirectory is the fallback location for markdown documentation files
var DocsDirectory = "./docs"

// FindDocContent returns the raw markdown documentation for a command,
// preferring the embedded docs and falling back to the local docs directory.
func FindDocContent(cmd *cobra.Command) ([]byte, error) {
	cmdPath := getCommandPath(cmd)
	docName := cmdPath + ".md"

	// embed.FS paths are always slash-separated, so use path.Join, not
	// filepath.Join (\ on Windows).
	if content, err := embeddedDocsFS.ReadFile(path.Join("docs", docName)); err == nil {
		return content, nil
	}

	// If embedded file not found, try local docs directory as fallback
	content, err := os.ReadFile(filepath.Join(DocsDirectory, docName))
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

	_, err = fmt.Fprintln(cmd.OutOrStdout(), rendered)
	return err
}
