package cmdbuilder

import (
	"embed"

	"github.com/spf13/cobra"
)

// embeddedDocsFS holds the embedded documentation files
var embeddedDocsFS embed.FS

// RegisterEmbeddedDocs registers the embedded documentation filesystem
func RegisterEmbeddedDocs(docs embed.FS) {
	embeddedDocsFS = docs
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
