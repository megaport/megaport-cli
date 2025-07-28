//go:build !js || !wasm
// +build !js !wasm

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

// main is the entry point for the application.
// For WASM builds, see main_wasm.go which defines the WASM-specific entry point.
func main() {
	// Register the embedded documentation with the cmdbuilder package
	cmdbuilder.RegisterEmbeddedDocs(embeddedDocs)

	megaport.Execute()
}
