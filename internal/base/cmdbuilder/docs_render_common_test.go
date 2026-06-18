package cmdbuilder

import (
	"embed"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/sample.md
var testDocsFS embed.FS

func TestGetCommandPath(t *testing.T) {
	root := &cobra.Command{Use: "megaport-cli"}
	mcr := &cobra.Command{Use: "mcr"}
	buy := &cobra.Command{Use: "buy"}
	root.AddCommand(mcr)
	mcr.AddCommand(buy)

	tests := []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{"root has no parent", root, "megaport-cli"},
		{"one level deep", mcr, "megaport-cli_mcr"},
		{"two levels deep", buy, "megaport-cli_mcr_buy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getCommandPath(tt.cmd))
		})
	}
}

func TestRegisterEmbeddedDocs(t *testing.T) {
	orig := embeddedDocsFS
	t.Cleanup(func() { embeddedDocsFS = orig })

	RegisterEmbeddedDocs(testDocsFS)

	content, err := embeddedDocsFS.ReadFile("testdata/sample.md")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Sample")
}
