package apply

import (
	"context"
	"fmt"
	"net/http"
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
	cmd.Flags().Bool("rollback-on-failure", false, "")
	_ = cmd.Flags().Set("file", file)
	if dryRun {
		_ = cmd.Flags().Set("dry-run", "true")
	}
	if yes {
		_ = cmd.Flags().Set("yes", "true")
	}
	return cmd
}

// applyCmdWithRollback builds a command with --rollback-on-failure enabled.
func applyCmdWithRollback(file string) *cobra.Command {
	cmd := applyCmd(file, false, true)
	_ = cmd.Flags().Set("rollback-on-failure", "true")
	return cmd
}

// setupMockClient overrides the login function with mock services and returns cleanup.
func setupMockClient(port *MockPortService, mcr *MockMCRService, mve *MockMVEService, vxc *MockVXCService) func() {
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = port
		client.MCRService = mcr
		client.MVEService = mve
		client.VXCService = vxc
		return client, nil
	})
	return func() { config.SetLoginFunc(original) }
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
	// Ordering no longer waits inline; apply tracks the UID then polls for readiness.
	assert.False(t, mockPort.CapturedPortRequest.WaitForProvision)
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
	apiErr := &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: 403,
			Header:     http.Header{},
			Request:    &http.Request{},
		},
		Message: "forbidden",
	}
	mockPort := &MockPortService{BuyPortErr: apiErr}
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
	assert.Contains(t, err.Error(), "permission denied")
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
	output.SetTerminalWidthForTesting(200)
	defer output.SetTerminalWidthForTesting(0)
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
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	})
	defer func() { config.SetLoginFunc(original) }()

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

func TestApplyConfig_MVEWithYAMLIntegerVendorConfig(t *testing.T) {
	// YAML decodes integer scalars as int (not float64 like JSON).
	// normalizeVendorConfigMap must convert them so ParseVendorConfig works.
	mockMVE := &MockMVEService{}
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, mockMVE, &MockVXCService{})()

	yaml := `
mves:
  - name: Test-MVE
    location_id: 1
    term: 1
    vendor_config:
      vendor: 6wind
      imageId: 42
      productSize: SMALL
      sshPublicKey: "ssh-rsa AAAA"
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	// Should not fail due to YAML int → float64 type mismatch in ParseVendorConfig
	require.NoError(t, err)
}

func TestApplyConfig_DryRunUnknownTemplateRef(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	// VXC references a port that is not declared in the config — dry-run should catch it.
	yaml := `
vxcs:
  - name: Orphan-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.Undeclared}}"
    b_end:
      product_uid: some-uid
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, true, false)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.NoError(t, err) // dry-run reports invalid, does not return error
	assert.Contains(t, captured, "invalid")
	assert.Contains(t, captured, "Undeclared")
}

func TestApplyConfig_MVEAPIError(t *testing.T) {
	apiErr := &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: 401,
			Header:     http.Header{},
			Request:    &http.Request{},
		},
		Message: "unauthorized",
	}
	mockMVE := &MockMVEService{BuyMVEErr: apiErr}
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, mockMVE, &MockVXCService{})()

	yaml := `
mves:
  - name: Bad-MVE
    location_id: 1
    term: 1
    vendor_config:
      vendor: 6wind
      imageId: 42
      productSize: SMALL
      sshPublicKey: "ssh-rsa AAAA"
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestApplyConfig_MCRAPIError(t *testing.T) {
	apiErr := &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: 429,
			Header:     http.Header{},
			Request:    &http.Request{},
		},
		Message: "rate limited",
	}
	mockMCR := &MockMCRService{BuyMCRErr: apiErr}
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
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestApplyConfig_VXCAPIError(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-uid-abc"}},
	}
	vxcAPIErr := &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: 401,
			Header:     http.Header{},
			Request:    &http.Request{},
		},
		Message: "unauthorized",
	}
	mockVXC := &MockVXCService{BuyVXCErr: vxcAPIErr}
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
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestApplyConfig_NoFileFlag(t *testing.T) {
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	cmd := &cobra.Command{Use: "apply"}
	cmd.Flags().StringP("file", "f", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	cmd.Flags().Bool("rollback-on-failure", false, "")
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

// --- orphan reporting and rollback tests ---

// TestApplyConfig_OrphanReporting verifies that when a port succeeds but the MCR
// fails, the output prominently lists the orphaned port UID and a delete command.
func TestApplyConfig_OrphanReporting(t *testing.T) {
	output.SetTerminalWidthForTesting(200)
	defer output.SetTerminalWidthForTesting(0)

	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-orphan-uid"}},
	}
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR API down")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
ports:
  - name: Orphan-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Failing-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmd(f, false, true)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCR API down")
	// Orphaned port must be reported with its UID and a remediation command.
	assert.Contains(t, captured, "port-orphan-uid")
	assert.Contains(t, captured, "megaport-cli ports delete port-orphan-uid")
	assert.Contains(t, captured, "ARE BILLING")
}

// TestApplyConfig_RollbackOnFailure verifies that with --rollback-on-failure, the
// port created before the MCR failure is deleted via DeletePort.
func TestApplyConfig_RollbackOnFailure(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-rollback-uid"}},
	}
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR API down")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
ports:
  - name: Rollback-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Failing-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCR API down")
	// DeletePort must have been called with the orphaned port's UID.
	require.Equal(t, []string{"port-rollback-uid"}, mockPort.DeletePortCalledWith)
}

// TestApplyConfig_RollbackOnFailure_DeleteError verifies that when rollback itself
// fails, the output instructs the user to delete manually.
func TestApplyConfig_RollbackOnFailure_DeleteError(t *testing.T) {
	output.SetTerminalWidthForTesting(200)
	defer output.SetTerminalWidthForTesting(0)

	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-stuck-uid"}},
		DeletePortErr: fmt.Errorf("delete also failed"),
	}
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR API down")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
ports:
  - name: Stuck-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Failing-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	// DeletePort was attempted.
	require.Equal(t, []string{"port-stuck-uid"}, mockPort.DeletePortCalledWith)
	// Manual remediation command must appear in output.
	assert.Contains(t, captured, "megaport-cli ports delete port-stuck-uid")
}

// TestApplyConfig_VXCFailOrphanReporting verifies that when both a port and MCR
// succeed but the VXC fails, both are reported as billing.
func TestApplyConfig_VXCFailOrphanReporting(t *testing.T) {
	output.SetTerminalWidthForTesting(200)
	defer output.SetTerminalWidthForTesting(0)

	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-vxcfail-uid"}},
	}
	mockMCR := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-vxcfail-uid"},
	}
	mockVXC := &MockVXCService{BuyVXCErr: fmt.Errorf("VXC quota exceeded")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, mockVXC)()

	cfg := `
ports:
  - name: VXCFail-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: VXCFail-MCR
    location_id: 1
    speed: 1000
    term: 12

vxcs:
  - name: Fail-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.port.VXCFail-Port}}"
    b_end:
      product_uid: "{{.mcr.VXCFail-MCR}}"
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmd(f, false, true)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "VXC quota exceeded")
	// Both the port and the MCR must appear in the billing warning.
	assert.Contains(t, captured, "port-vxcfail-uid")
	assert.Contains(t, captured, "megaport-cli ports delete port-vxcfail-uid")
	assert.Contains(t, captured, "mcr-vxcfail-uid")
	assert.Contains(t, captured, "megaport-cli mcr delete mcr-vxcfail-uid")
}

// TestApplyConfig_RollbackOnFailure_MCR verifies that with --rollback-on-failure,
// an MCR created before a VXC failure is deleted via DeleteMCR.
func TestApplyConfig_RollbackOnFailure_MCR(t *testing.T) {
	mockMCR := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-rollback-uid"},
	}
	mockVXC := &MockVXCService{BuyVXCErr: fmt.Errorf("VXC quota exceeded")}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, mockVXC)()

	cfg := `
mcrs:
  - name: Rollback-MCR
    location_id: 1
    speed: 1000
    term: 12

vxcs:
  - name: Fail-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.mcr.Rollback-MCR}}"
    b_end:
      product_uid: "ext-uid"
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "VXC quota exceeded")
	require.Equal(t, []string{"mcr-rollback-uid"}, mockMCR.DeleteMCRCalledWith)
}

// TestApplyConfig_RollbackOnFailure_MVE verifies that with --rollback-on-failure,
// an MVE created before a VXC failure is deleted via DeleteMVE.
func TestApplyConfig_RollbackOnFailure_MVE(t *testing.T) {
	mockMVE := &MockMVEService{
		BuyMVEResult: &megaport.BuyMVEResponse{TechnicalServiceUID: "mve-rollback-uid"},
	}
	mockVXC := &MockVXCService{BuyVXCErr: fmt.Errorf("VXC quota exceeded")}
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, mockMVE, mockVXC)()

	cfg := `
mves:
  - name: Rollback-MVE
    location_id: 1
    term: 12
    vendor_config:
      vendor: 6wind
      imageId: 42
      productSize: SMALL
      sshPublicKey: "ssh-rsa AAAA"

vxcs:
  - name: Fail-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.mve.Rollback-MVE}}"
    b_end:
      product_uid: "ext-uid"
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "VXC quota exceeded")
	require.Equal(t, []string{"mve-rollback-uid"}, mockMVE.DeleteMVECalledWith)
}

// TestApplyConfig_RollbackOnFailure_VXC verifies that with --rollback-on-failure,
// a successfully created VXC is deleted when a subsequent VXC fails.
func TestApplyConfig_RollbackOnFailure_VXC(t *testing.T) {
	mockMCR := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-vxcroll-uid"},
	}
	// Second BuyVXC call fails; first succeeds with default UID "vxc-uid-mock-1".
	mockVXC := &MockVXCService{
		BuyVXCErr:       fmt.Errorf("second VXC failed"),
		BuyVXCErrOnCall: 2,
	}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, mockVXC)()

	cfg := `
mcrs:
  - name: VXCRoll-MCR
    location_id: 1
    speed: 1000
    term: 12

vxcs:
  - name: First-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.mcr.VXCRoll-MCR}}"
    b_end:
      product_uid: "ext-uid-1"
  - name: Second-VXC
    rate_limit: 100
    term: 12
    a_end:
      product_uid: "{{.mcr.VXCRoll-MCR}}"
    b_end:
      product_uid: "ext-uid-2"
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "second VXC failed")
	// First VXC and the MCR must both be rolled back.
	require.Equal(t, []string{"vxc-uid-mock-1"}, mockVXC.DeleteVXCCalledWith)
	require.Equal(t, []string{"mcr-vxcroll-uid"}, mockMCR.DeleteMCRCalledWith)
}

// TestApplyConfig_RollbackOnFailure_ProvisionTimeout verifies that a resource whose
// order succeeds but whose provision wait fails is tracked and rolled back. This is the
// orphan window the no-wait-then-poll restructure closes: the order has placed billing
// before provisioning completes, so the UID must already be recorded when the wait fails.
func TestApplyConfig_RollbackOnFailure_ProvisionTimeout(t *testing.T) {
	mockMCR := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-provision-fail-uid"},
		GetMCRErr:    fmt.Errorf("provisioning status check failed"),
	}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
mcrs:
  - name: Provision-Fail-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "table")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provisioning status check failed")
	// The order placed billing before provisioning completed, so the MCR must be
	// rolled back even though it never reached a ready state.
	require.Equal(t, []string{"mcr-provision-fail-uid"}, mockMCR.DeleteMCRCalledWith)
}

// TestApplyConfig_RollbackOnFailure_JSONMode verifies that in --output json mode with
// --rollback-on-failure, both successful and failed rollbacks appear in the returned
// error so JSON consumers get a complete picture of what was cleaned up.
func TestApplyConfig_RollbackOnFailure_JSONMode(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-json-roll-uid"}},
	}
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR API down")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
ports:
  - name: JSON-Roll-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Failing-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmdWithRollback(f)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "json")
	})

	require.Error(t, err)
	// Successful rollback must appear in the error for JSON consumers.
	assert.Contains(t, err.Error(), "rolled back")
	assert.Contains(t, err.Error(), "port-json-roll-uid")
	// No plain-text rollback lines in captured output.
	assert.NotContains(t, captured, "Rolled back")
}

// TestApplyConfig_OrphanReporting_JSONMode verifies that in --output json mode the
// orphan details are embedded in the returned error (not emitted as plain-text lines)
// so the JSON error envelope the wrapper prints is the only structured output on stderr.
func TestApplyConfig_OrphanReporting_JSONMode(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-json-uid"}},
	}
	mockMCR := &MockMCRService{BuyMCRErr: fmt.Errorf("MCR API down")}
	defer setupMockClient(mockPort, mockMCR, &MockMVEService{}, &MockVXCService{})()

	cfg := `
ports:
  - name: JSON-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

mcrs:
  - name: Failing-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", cfg)
	cmd := applyCmd(f, false, true)

	var err error
	captured := output.CaptureOutput(func() {
		err = ApplyConfig(cmd, nil, true, "json")
	})

	require.Error(t, err)
	// Orphan details must be in the error, not in captured stdout/stderr text.
	assert.Contains(t, err.Error(), "port-json-uid")
	assert.Contains(t, err.Error(), "ARE BILLING")
	assert.Contains(t, err.Error(), "megaport-cli ports delete port-json-uid")
	// No plain-text billing lines should appear in the captured output.
	assert.NotContains(t, captured, "ARE BILLING")
}

// TestApplyModule_RegistersRollbackFlag checks that AddCommandsTo registers the
// rollback-on-failure flag so Cobra exposes it to users.
func TestApplyModule_RegistersRollbackFlag(t *testing.T) {
	m := &Module{}
	root := &cobra.Command{Use: "megaport-cli"}
	m.RegisterCommands(root)
	require.Len(t, root.Commands(), 1)
	applyC := root.Commands()[0]
	assert.NotNil(t, applyC.Flag("rollback-on-failure"))
}

// --- nil API response tests ---

func TestApplyConfig_PortNilResponse(t *testing.T) {
	mockPort := &MockPortService{BuyPortNilResp: true}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, &MockVXCService{})()

	yaml := `
ports:
  - name: Nil-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	require.NotPanics(t, func() {
		output.CaptureOutput(func() {
			err = ApplyConfig(cmd, nil, true, "table")
		})
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

func TestApplyConfig_MCRNilResponse(t *testing.T) {
	mockMCR := &MockMCRService{BuyMCRNilResp: true}
	defer setupMockClient(&MockPortService{}, mockMCR, &MockMVEService{}, &MockVXCService{})()

	yaml := `
mcrs:
  - name: Nil-MCR
    location_id: 1
    speed: 1000
    term: 12
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	require.NotPanics(t, func() {
		output.CaptureOutput(func() {
			err = ApplyConfig(cmd, nil, true, "table")
		})
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

func TestApplyConfig_MVENilResponse(t *testing.T) {
	mockMVE := &MockMVEService{BuyMVENilResp: true}
	defer setupMockClient(&MockPortService{}, &MockMCRService{}, mockMVE, &MockVXCService{})()

	yaml := `
mves:
  - name: Nil-MVE
    location_id: 1
    term: 1
    vendor_config:
      vendor: 6wind
      imageId: 42
      productSize: SMALL
      sshPublicKey: "ssh-rsa AAAA"
`
	f := writeTempFile(t, "config.yaml", yaml)
	cmd := applyCmd(f, false, true)

	var err error
	require.NotPanics(t, func() {
		output.CaptureOutput(func() {
			err = ApplyConfig(cmd, nil, true, "table")
		})
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

func TestApplyConfig_VXCNilResponse(t *testing.T) {
	mockPort := &MockPortService{
		BuyPortResult: &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-uid-abc"}},
	}
	mockVXC := &MockVXCService{BuyVXCNilResp: true}
	defer setupMockClient(mockPort, &MockMCRService{}, &MockMVEService{}, mockVXC)()

	yaml := `
ports:
  - name: Test-Port
    location_id: 1
    speed: 1000
    term: 12
    marketplace_visibility: false

vxcs:
  - name: Nil-VXC
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
	require.NotPanics(t, func() {
		output.CaptureOutput(func() {
			err = ApplyConfig(cmd, nil, true, "table")
		})
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
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

func TestNormalizeVendorConfigMap_IntToFloat64(t *testing.T) {
	// YAML decodes integers as int; normalizeVendorConfigMap converts them to float64
	// so that ParseVendorConfig (which uses float64 type assertions) works correctly.
	input := map[string]interface{}{
		"vendor":      "cisco",
		"imageId":     42, // int from YAML
		"productSize": "SMALL",
	}
	out, err := normalizeVendorConfigMap(input)
	require.NoError(t, err)
	// After round-trip through JSON, integers become float64
	assert.IsType(t, float64(0), out["imageId"])
	assert.Equal(t, float64(42), out["imageId"])
	assert.Equal(t, "cisco", out["vendor"])
}

func TestNormalizeVendorConfigMap_Nil(t *testing.T) {
	out, err := normalizeVendorConfigMap(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
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
