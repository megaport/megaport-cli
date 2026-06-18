//go:build !js || !wasm

package cmdbuilder

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// ShowDocumentation must write to the command's configured output writer, not
// straight to os.Stdout, so callers (and tests) can capture/redirect docs output.
func TestShowDocumentationWritesToCobraWriter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "megaport-cli.md"), []byte("# Title\n\nbody text\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir := DocsDirectory
	DocsDirectory = dir
	t.Cleanup(func() { DocsDirectory = origDir })

	origFS := embeddedDocsFS
	embeddedDocsFS = embed.FS{} // force the on-disk fallback
	t.Cleanup(func() { embeddedDocsFS = origFS })

	root := &cobra.Command{Use: "megaport-cli"}
	var buf bytes.Buffer
	root.SetOut(&buf)

	if err := ShowDocumentation(root); err != nil {
		t.Fatalf("ShowDocumentation: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected rendered docs on the cobra writer, got nothing")
	}
}
