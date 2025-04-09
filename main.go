package main

import (
	"embed"

	"github.com/megaport/megaport-cli/cmd/megaport"
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
)

// Embed all documentation files into the binary
//
//go:embed docs/*.md
var embeddedDocs embed.FS

func main() {
	// Register the embedded documentation with the cmdbuilder package
	cmdbuilder.RegisterEmbeddedDocs(embeddedDocs)

	megaport.Execute()
}
