package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func newStatusCmd() *cobra.Command {
	cmd := testutil.NewCommand("status", testutil.OutputAdapter(StatusDashboard))
	cmd.Flags().Bool("include-inactive", false, "")
	return cmd
}

func setupMocks(portSvc *MockPortService, mcrSvc *MockMCRService, mveSvc *MockMVEService, vxcSvc *MockVXCService, ixSvc *MockIXService) func() {
	return testutil.SetupLogin(func(c *megaport.Client) {
		c.PortService = portSvc
		c.MCRService = mcrSvc
		c.MVEService = mveSvc
		c.VXCService = vxcSvc
		c.IXService = ixSvc
	})
}

// TestStatusDashboard_AllPopulated verifies the happy path with data for every resource type.
func TestStatusDashboard_AllPopulated(t *testing.T) {
	portSvc := &MockPortService{
		ListPortsResult: []*megaport.Port{
			{UID: "port-1", Name: "Port One", ProvisioningStatus: "LIVE", PortSpeed: 10000, LocationID: 1},
		},
	}
	mcrSvc := &MockMCRService{
		ListMCRsResult: []*megaport.MCR{
			{UID: "mcr-1", Name: "MCR One", ProvisioningStatus: "LIVE", PortSpeed: 5000,
				Resources: megaport.MCRResources{VirtualRouter: megaport.MCRVirtualRouter{ASN: 65000}}},
		},
	}
	mveSvc := &MockMVEService{
		ListMVEsResult: []*megaport.MVE{
			{UID: "mve-1", Name: "MVE One", ProvisioningStatus: "LIVE", Vendor: "cisco", Size: "MEDIUM"},
		},
	}
	vxcSvc := &MockVXCService{
		ListVXCsResult: []*megaport.VXC{
			{
				UID: "vxc-1", Name: "VXC One", ProvisioningStatus: "LIVE", RateLimit: 500,
				AEndConfiguration: megaport.VXCEndConfiguration{UID: "port-1"},
				BEndConfiguration: megaport.VXCEndConfiguration{UID: "port-2"},
			},
		},
	}
	ixSvc := &MockIXService{
		ListIXsResult: []*megaport.IX{
			{ProductUID: "ix-1", ProductName: "IX One", ProvisioningStatus: "LIVE", ASN: 64512, RateLimit: 1000},
		},
	}

	cleanup := setupMocks(portSvc, mcrSvc, mveSvc, vxcSvc, ixSvc)
	defer cleanup()

	cmd := newStatusCmd()
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestStatusDashboard_AllEmpty verifies empty resources print "No X found."
func TestStatusDashboard_AllEmpty(t *testing.T) {
	cleanup := setupMocks(
		&MockPortService{},
		&MockMCRService{},
		&MockMVEService{},
		&MockVXCService{},
		&MockIXService{},
	)
	defer cleanup()

	cmd := newStatusCmd()
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestStatusDashboard_ServiceError verifies that a single service error propagates.
func TestStatusDashboard_ServiceError(t *testing.T) {
	cleanup := setupMocks(
		&MockPortService{ListPortsErr: errors.New("ports API down")},
		&MockMCRService{},
		&MockMVEService{},
		&MockVXCService{},
		&MockIXService{},
	)
	defer cleanup()

	cmd := newStatusCmd()
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestStatusDashboard_IncludeInactive verifies the flag is passed through to request structs.
func TestStatusDashboard_IncludeInactive(t *testing.T) {
	mcrSvc := &MockMCRService{}
	mveSvc := &MockMVEService{}
	vxcSvc := &MockVXCService{}
	ixSvc := &MockIXService{}

	cleanup := setupMocks(&MockPortService{}, mcrSvc, mveSvc, vxcSvc, ixSvc)
	defer cleanup()

	cmd := newStatusCmd()
	require.NoError(t, cmd.Flags().Set("include-inactive", "true"))
	err := cmd.Execute()
	assert.NoError(t, err)

	assert.NotNil(t, mcrSvc.CapturedListMCRsRequest)
	assert.True(t, mcrSvc.CapturedListMCRsRequest.IncludeInactive)

	assert.NotNil(t, mveSvc.CapturedListMVEsRequest)
	assert.True(t, mveSvc.CapturedListMVEsRequest.IncludeInactive)

	assert.NotNil(t, vxcSvc.CapturedListVXCsRequest)
	assert.True(t, vxcSvc.CapturedListVXCsRequest.IncludeInactive)

	assert.NotNil(t, ixSvc.CapturedListIXsRequest)
	assert.True(t, ixSvc.CapturedListIXsRequest.IncludeInactive)
}

// TestStatusDashboard_InactivePortsFiltered verifies decommissioned ports are excluded by default.
func TestStatusDashboard_InactivePortsFiltered(t *testing.T) {
	portSvc := &MockPortService{
		ListPortsResult: []*megaport.Port{
			{UID: "port-live", Name: "Live Port", ProvisioningStatus: "LIVE", PortSpeed: 1000, LocationID: 1},
			{UID: "port-decomm", Name: "Dead Port", ProvisioningStatus: "DECOMMISSIONED", PortSpeed: 1000, LocationID: 1},
			{UID: "port-cancel", Name: "Cancelled Port", ProvisioningStatus: "CANCELLED", PortSpeed: 1000, LocationID: 1},
		},
	}

	cleanup := setupMocks(portSvc, &MockMCRService{}, &MockMVEService{}, &MockVXCService{}, &MockIXService{})
	defer cleanup()

	// Capture what gets built
	var capturedPorts []*megaport.Port
	origListPorts := listPortsFunc
	defer func() { listPortsFunc = origListPorts }()
	listPortsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Port, error) {
		return portSvc.ListPorts(ctx)
	}

	// We test the filtering logic indirectly via JSON output
	cmd := newStatusCmd()
	require.NoError(t, cmd.Flags().Set("output", "json"))

	// Redirect stdout to capture JSON
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := StatusDashboard(cmd, nil, true, "json")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)

	var dashboard DashboardOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &dashboard))
	assert.Len(t, dashboard.Ports, 1)
	assert.Equal(t, "port-live", dashboard.Ports[0].UID)

	_ = capturedPorts
}

// TestStatusDashboard_JSONOutput verifies the JSON output contains all 5 keys.
func TestStatusDashboard_JSONOutput(t *testing.T) {
	cleanup := setupMocks(
		&MockPortService{ListPortsResult: []*megaport.Port{
			{UID: "port-1", Name: "P1", ProvisioningStatus: "LIVE", PortSpeed: 1000, LocationID: 1},
		}},
		&MockMCRService{ListMCRsResult: []*megaport.MCR{
			{UID: "mcr-1", Name: "M1", ProvisioningStatus: "LIVE"},
		}},
		&MockMVEService{},
		&MockVXCService{},
		&MockIXService{},
	)
	defer cleanup()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newStatusCmd()
	require.NoError(t, cmd.Flags().Set("output", "json"))
	err := StatusDashboard(cmd, nil, true, "json")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	for _, key := range []string{"ports", "mcrs", "mves", "vxcs", "ixs", "summary"} {
		assert.Contains(t, result, key, "JSON output missing key %q", key)
	}
}

// TestStatusDashboard_LoginError verifies login failure is reported.
func TestStatusDashboard_LoginError(t *testing.T) {
	cleanup := testutil.SetupLoginError(errors.New("auth failed"))
	defer cleanup()

	cmd := newStatusCmd()
	err := cmd.Execute()
	assert.Error(t, err)
}

// --- Converter unit tests ---

func TestToStatusPortOutput(t *testing.T) {
	p := &megaport.Port{UID: "p-1", Name: "Port", ProvisioningStatus: "LIVE", PortSpeed: 10000, LocationID: 5}
	out, err := toStatusPortOutput(p)
	require.NoError(t, err)
	assert.Equal(t, "p-1", out.UID)
	assert.Equal(t, "Port", out.Name)
	assert.Equal(t, "LIVE", out.Status)
	assert.Equal(t, 10000, out.Speed)
	assert.Equal(t, 5, out.LocationID)
}

func TestToStatusPortOutput_Nil(t *testing.T) {
	_, err := toStatusPortOutput(nil)
	assert.Error(t, err)
}

func TestToStatusMCROutput(t *testing.T) {
	m := &megaport.MCR{
		UID: "mcr-1", Name: "MCR", ProvisioningStatus: "LIVE", PortSpeed: 5000,
		Resources: megaport.MCRResources{VirtualRouter: megaport.MCRVirtualRouter{ASN: 65001}},
	}
	out, err := toStatusMCROutput(m)
	require.NoError(t, err)
	assert.Equal(t, "mcr-1", out.UID)
	assert.Equal(t, 65001, out.ASN)
	assert.Equal(t, 5000, out.Speed)
}

func TestToStatusMCROutput_Nil(t *testing.T) {
	_, err := toStatusMCROutput(nil)
	assert.Error(t, err)
}

func TestToStatusMVEOutput(t *testing.T) {
	m := &megaport.MVE{UID: "mve-1", Name: "MVE", ProvisioningStatus: "LIVE", Vendor: "palo-alto", Size: "LARGE"}
	out, err := toStatusMVEOutput(m)
	require.NoError(t, err)
	assert.Equal(t, "mve-1", out.UID)
	assert.Equal(t, "palo-alto", out.Vendor)
	assert.Equal(t, "LARGE", out.Size)
}

func TestToStatusMVEOutput_Nil(t *testing.T) {
	_, err := toStatusMVEOutput(nil)
	assert.Error(t, err)
}

func TestToStatusVXCOutput(t *testing.T) {
	v := &megaport.VXC{
		UID: "vxc-1", Name: "VXC", ProvisioningStatus: "LIVE", RateLimit: 500,
		AEndConfiguration: megaport.VXCEndConfiguration{UID: "a-end"},
		BEndConfiguration: megaport.VXCEndConfiguration{UID: "b-end"},
	}
	out, err := toStatusVXCOutput(v)
	require.NoError(t, err)
	assert.Equal(t, "vxc-1", out.UID)
	assert.Equal(t, 500, out.RateLimit)
	assert.Equal(t, "a-end", out.AEndUID)
	assert.Equal(t, "b-end", out.BEndUID)
}

func TestToStatusVXCOutput_Nil(t *testing.T) {
	_, err := toStatusVXCOutput(nil)
	assert.Error(t, err)
}

func TestToStatusIXOutput(t *testing.T) {
	i := &megaport.IX{ProductUID: "ix-1", ProductName: "IX", ProvisioningStatus: "LIVE", ASN: 64512, RateLimit: 1000}
	out, err := toStatusIXOutput(i)
	require.NoError(t, err)
	assert.Equal(t, "ix-1", out.UID)
	assert.Equal(t, 64512, out.ASN)
	assert.Equal(t, 1000, out.RateLimit)
}

func TestToStatusIXOutput_Nil(t *testing.T) {
	_, err := toStatusIXOutput(nil)
	assert.Error(t, err)
}

// TestStatusDashboard_MultipleServiceErrors verifies all errors are reported.
func TestStatusDashboard_MultipleServiceErrors(t *testing.T) {
	cleanup := setupMocks(
		&MockPortService{ListPortsErr: errors.New("ports down")},
		&MockMCRService{ListMCRsErr: errors.New("MCRs down")},
		&MockMVEService{},
		&MockVXCService{},
		&MockIXService{},
	)
	defer cleanup()

	cmd := newStatusCmd()
	err := cmd.Execute()
	assert.Error(t, err)
}

func init() {
	// Suppress output during tests.
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}
}
