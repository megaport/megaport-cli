//go:build js && wasm

package cmdbuilder

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ShowDocumentation prints the embedded markdown for a command. The browser
// terminal styles it, so glamour (and its goldmark/chroma deps) is left out of
// the WASM build.
func ShowDocumentation(cmd *cobra.Command) error {
	cmdPath := getCommandPath(cmd)

	content, err := embeddedDocsFS.ReadFile(filepath.Join("docs", cmdPath+".md"))
	if err != nil {
		return fmt.Errorf("documentation file not found for %s: %w", cmdPath, err)
	}

	fmt.Println(string(content))
	return nil
}
