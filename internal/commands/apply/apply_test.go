package apply

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTemplates_NoTemplate(t *testing.T) {
	uids := map[string]map[string]string{"port": {"Sydney": "uid-abc"}}
	result, err := resolveTemplates("some-fixed-uid", uids)
	require.NoError(t, err)
	assert.Equal(t, "some-fixed-uid", result)
}

func TestResolveTemplates_ResolvesReference(t *testing.T) {
	uids := map[string]map[string]string{
		"port": {"Sydney-Primary": "port-uid-123"},
		"mcr":  {"Sydney-MCR": "mcr-uid-456"},
	}

	result, err := resolveTemplates("{{.port.Sydney-Primary}}", uids)
	require.NoError(t, err)
	assert.Equal(t, "port-uid-123", result)

	result, err = resolveTemplates("{{.mcr.Sydney-MCR}}", uids)
	require.NoError(t, err)
	assert.Equal(t, "mcr-uid-456", result)
}

func TestResolveTemplates_UnknownReference(t *testing.T) {
	uids := map[string]map[string]string{"port": {}}
	_, err := resolveTemplates("{{.port.Missing}}", uids)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing")
}

func TestResolveTemplates_UnknownType(t *testing.T) {
	uids := map[string]map[string]string{}
	_, err := resolveTemplates("{{.vxc.SomeVXC}}", uids)
	assert.Error(t, err)
}

func TestResolveOrPlaceholder_NoTemplate(t *testing.T) {
	assert.Equal(t, "real-uid", resolveOrPlaceholder("real-uid"))
}

func TestResolveOrPlaceholder_WithTemplate(t *testing.T) {
	result := resolveOrPlaceholder("{{.port.Sydney}}")
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", result)
}

func TestParseConfigFile_YAML(t *testing.T) {
	content := `
ports:
  - name: Test-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: true

mcrs:
  - name: Test-MCR
    location_id: 2
    speed: 1000
    term: 12
    asn: 65000

vxcs:
  - name: Test-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.Test-Port}}"
      vlan: 100
    b_end:
      product_uid: "{{.mcr.Test-MCR}}"
`
	f := writeTempFile(t, "config.yaml", content)
	cfg, err := parseConfigFile(f)
	require.NoError(t, err)

	require.Len(t, cfg.Ports, 1)
	assert.Equal(t, "Test-Port", cfg.Ports[0].Name)
	assert.Equal(t, 1000, cfg.Ports[0].Speed)
	assert.True(t, cfg.Ports[0].MarketplaceVisibility)

	require.Len(t, cfg.MCRs, 1)
	assert.Equal(t, "Test-MCR", cfg.MCRs[0].Name)
	assert.Equal(t, 65000, cfg.MCRs[0].ASN)

	require.Len(t, cfg.VXCs, 1)
	assert.Equal(t, "{{.port.Test-Port}}", cfg.VXCs[0].AEnd.ProductUID)
	assert.Equal(t, 100, cfg.VXCs[0].AEnd.VLAN)
}

func TestParseConfigFile_JSON(t *testing.T) {
	content := `{
  "ports": [
    {"name": "JSON-Port", "location_id": 5, "speed": 10000, "term": 24}
  ]
}`
	f := writeTempFile(t, "config.json", content)
	cfg, err := parseConfigFile(f)
	require.NoError(t, err)

	require.Len(t, cfg.Ports, 1)
	assert.Equal(t, "JSON-Port", cfg.Ports[0].Name)
	assert.Equal(t, 10000, cfg.Ports[0].Speed)
}

func TestParseConfigFile_NotFound(t *testing.T) {
	_, err := parseConfigFile("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestParseConfigFile_InvalidYAML(t *testing.T) {
	f := writeTempFile(t, "bad.yaml", "ports: [invalid yaml }")
	_, err := parseConfigFile(f)
	assert.Error(t, err)
}

func TestParseConfigFile_EmptyFile(t *testing.T) {
	f := writeTempFile(t, "empty.yaml", "")
	cfg, err := parseConfigFile(f)
	require.NoError(t, err)
	assert.Empty(t, cfg.Ports)
	assert.Empty(t, cfg.MCRs)
	assert.Empty(t, cfg.VXCs)
}

// writeTempFile creates a temp file with the given name suffix and content,
// returning its path.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}
