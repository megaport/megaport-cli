package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// applyCmd builds a minimal cobra.Command with the flags ApplyConfig reads.
func applyCmd(file string, dryRun, yes bool) *cobra.Command {
	cmd := &cobra.Command{Use: "apply"}
	cmd.Flags().StringP("file", "f", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	_ = cmd.Flags().Set("file", file)
	if dryRun {
		_ = cmd.Flags().Set("dry-run", "true")
	}
	if yes {
		_ = cmd.Flags().Set("yes", "true")
	}
	return cmd
}

// setupMockClient overrides config.LoginFunc with mock services and returns cleanup.
func setupMockClient(port *MockPortService, mcr *MockMCRService, mve *MockMVEService, vxc *MockVXCService) func() {
	original := config.LoginFunc
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = port
		client.MCRService = mcr
		client.MVEService = mve
		client.VXCService = vxc
		return client, nil
	}
	return func() { config.LoginFunc = original }
}

func TestApplyConfig_EmptyConfig(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	f := writeTempFile(t, "empty.yaml", "")
	cmd := applyCmd(f, false, true)

	output.CaptureOutput(func() {
		err := ApplyConfig(cmd, nil, true, "table")
		assert.NoError(t, err)
	})
}

func TestApplyConfig_ProvisionPort(t *testing.T) {
	mockPort := &MockPortService{}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
ports:
  - name: Test-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: true
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	require.NotNil(t, mockPort.CapturedPortRequest)
	assert.Equal(t, "Test-Port", mockPort.CapturedPortRequest.Name)
	assert.Equal(t, 1, mockPort.CapturedPortRequest.LocationId)
	assert.Equal(t, 1000, mockPort.CapturedPortRequest.PortSpeed)
	assert.Equal(t, 12, mockPort.CapturedPortRequest.Term)
	assert.True(t, mockPort.CapturedPortRequest.WaitForProvision)
}

func TestApplyConfig_ProvisionMCR(t *testing.T) {
	mockMCR := &MockMCRService{}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, &MockVXCService{})()

	yaml := `
mcrs:
  - name: Test-MCR
    location_id: 2
    speed: 1000
    term: 12
    asn: 65000
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	require.NotNil(t, mockMCR.CapturedMCRRequest)
	assert.Equal(t, "Test-MCR", mockMCR.CapturedMCRRequest.Name)
	assert.Equal(t, 65000, mockMCR.CapturedMCRRequest.MCRAsn)
}

func TestApplyConfig_ProvisionVXCWithTemplateRef(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-uid-abc"}},
	}
	mockMCR := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-uid-def"},
	}
	mockVXC := &MockVXCService{}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, mockVXC)()

	yaml := `
ports:
  - name: Sydney-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Sydney-MCR
    location_id: 1
    speed: 1000
    term: 12

vxcs:
  - name: Port-to-MCR
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.Sydney-Port}}"
      vlan: 100
    b_end:
      product_uid: "{{.mcr.Sydney-MCR}}"
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	require.NotNil(t, mockVXC.CapturedVXCRequest)
	assert.Equal(t, "port-uid-abc", mockVXC.CapturedVXCRequest.AEndConfiguration.ProductUID)
	assert.Equal(t, "mcr-uid-def", mockVXC.CapturedVXCRequest.BEndConfiguration.ProductUID)
	assert.Equal(t, 100, mockVXC.CapturedVXCRequest.AEndConfiguration.VLAN)
}

func TestApplyConfig_PortAPIError(t *testing.T) {
	mockPort := &MockPortService{BuyPortErr: fmt.Errorf("API unavailable")}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
ports:
  - name: Bad-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API unavailable")
}

func TestApplyConfig_UnresolvedTemplate(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
vxcs:
  - name: Bad-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.Nonexistent}}"
    b_end:
      product_uid: some-uid
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Nonexistent")
}

func TestApplyConfig_DryRunPorts(t *testing.T) {
	mockPort := &MockPortService{}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
ports:
  - name: DryRun-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, true, false)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	// Dry-run should NOT call BuyPort
	assert.Nil(t, mockPort.CapturedPortRequest)
	assert.Contains(t, captured, "DryRun-Port")
}

func TestApplyConfig_DryRunValidationError(t *testing.T) {
	mockPort := &MockPortService{ValidatePortOrderErr: fmt.Errorf("invalid location")}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
ports:
  - name: Invalid-Port
    location_id: 9999
    speed: 1000
    term: 12
    marketplace_visibility: false
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, true, false)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	// Dry-run reports validation errors in the table but does not return an error
	require.NoError(t, err)
	assert.Contains(t, captured, "invalid location")
}

func TestApplyConfig_MissingFile(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	cmd := applyCmd("/nonexistent/path.yaml", false, true)
	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
}

func TestApplyConfig_LoginError(t *testing.T) {
	original := config.LoginFunc
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	}
	defer func() { config.LoginFunc = original }()

	f := writeTempFile(t, "empty.yaml", "")
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestApplyConfig_JSONFormat(t *testing.T) {
	mockPort := &MockPortService{}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	content := `{"ports":[{"name":"JSON-Port","location_id":1,"speed":1000,"term":12}]}`
	f := writeTempFile(t, "config.json", content)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "json")
	})

	require.NoError(t, err)
	require.NotNil(t, mockPort.CapturedPortRequest)
	assert.Equal(t, "JSON-Port", mockPort.CapturedPortRequest.Name)
}

func TestApplyConfig_MCRAPIError(t *testing.T) {
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR unavailable")}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, &MockVXCService{})()

	yaml := `
mcrs:
  - name: Bad-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCR unavailable")
}

func TestApplyConfig_VXCAPIError(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-uid-abc"}},
	}
	mockVXC := &MockVXCService{BuyVXCErr: fmt.Errorf("VXC service down")}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, mockVXC)()

	yaml := `
ports:
  - name: Test-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

vxcs:
  - name: Bad-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.Test-Port}}"
    b_end:
      product_uid: some-uid
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "VXC service down")
}

func TestApplyConfig_NoFileFlag(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	cmd := &cobra.Command{Use: "apply"}
	cmd.Flags().StringP("file", "f", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	// file flag intentionally not set

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--file is required")
}

func TestApplyConfig_DryRunMCRAndVXC(t *testing.T) {
	mockMCR := &MockMCRService{}
	mockVXC := &MockVXCService{}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, mockVXC)()

	yaml := `
mcrs:
  - name: Dry-MCR
    location_id: 1
    speed: 1000
    term: 12

vxcs:
  - name: Dry-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.SomePort}}"
    b_end:
      product_uid: some-uid
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, true, false)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	assert.Contains(t, captured, "Dry-MCR")
	assert.Contains(t, captured, "Dry-VXC")
	// BuyMCR/BuyVXC should NOT have been called
	assert.Nil(t, mockMCR.CapturedMCRRequest)
	assert.Nil(t, mockVXC.CapturedVXCRequest)
}

func TestApplyConfig_DryRunMVEInvalidVendorConfig(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
mves:
  - name: Bad-MVE
    location_id: 1
    term: 12
    vendor_config:
      vendor: unknown_vendor
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, true, false)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err)
	assert.Contains(t, captured, "invalid")
}

// --- helpers (also used by other test functions) ---

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
