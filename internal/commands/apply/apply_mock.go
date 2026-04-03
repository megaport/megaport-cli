package apply

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// MockPortService implements megaport.PortService for testing.
type MockPortService struct {
	BuyPortResult        *megaport.BuyPortResponse
	BuyPortErr           error
	ValidatePortOrderErr error
	CapturedPortRequest  *megaport.BuyPortRequest
}

func (m *MockPortService) BuyPort(ctx context.Context, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	m.CapturedPortRequest = req
	if m.BuyPortErr != nil {
		return nil, m.BuyPortErr
	}
	if m.BuyPortResult != nil {
		return m.BuyPortResult, nil
	}
	return &megaport.BuyPortResponse{TechnicalServiceUIDs: []string{"port-uid-mock"}}, nil
}

func (m *MockPortService) ValidatePortOrder(ctx context.Context, req *megaport.BuyPortRequest) error {
	return m.ValidatePortOrderErr
}

func (m *MockPortService) ListPorts(ctx context.Context) ([]*megaport.Port, error) { return nil, nil }
func (m *MockPortService) GetPort(ctx context.Context, portId string) (*megaport.Port, error) {
	return nil, nil
}
func (m *MockPortService) ModifyPort(ctx context.Context, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return nil, nil
}
func (m *MockPortService) DeletePort(ctx context.Context, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return nil, nil
}
func (m *MockPortService) RestorePort(ctx context.Context, portId string) (*megaport.RestorePortResponse, error) {
	return nil, nil
}
func (m *MockPortService) LockPort(ctx context.Context, portId string) (*megaport.LockPortResponse, error) {
	return nil, nil
}
func (m *MockPortService) UnlockPort(ctx context.Context, portId string) (*megaport.UnlockPortResponse, error) {
	return nil, nil
}
func (m *MockPortService) CheckPortVLANAvailability(ctx context.Context, portId string, vlan int) (bool, error) {
	return true, nil
}
func (m *MockPortService) ListPortResourceTags(ctx context.Context, portID string) (map[string]string, error) {
	return nil, nil
}
func (m *MockPortService) UpdatePortResourceTags(ctx context.Context, portID string, tags map[string]string) error {
	return nil
}

// MockMCRService implements megaport.MCRService for testing.
type MockMCRService struct {
	BuyMCRResult        *megaport.BuyMCRResponse
	BuyMCRErr           error
	ValidateMCROrderErr error
	CapturedMCRRequest  *megaport.BuyMCRRequest
}

func (m *MockMCRService) BuyMCR(ctx context.Context, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	m.CapturedMCRRequest = req
	if m.BuyMCRErr != nil {
		return nil, m.BuyMCRErr
	}
	if m.BuyMCRResult != nil {
		return m.BuyMCRResult, nil
	}
	return &megaport.BuyMCRResponse{TechnicalServiceUID: "mcr-uid-mock"}, nil
}

func (m *MockMCRService) ValidateMCROrder(ctx context.Context, req *megaport.BuyMCRRequest) error {
	return m.ValidateMCROrderErr
}

func (m *MockMCRService) ListMCRs(ctx context.Context, req *megaport.ListMCRsRequest) ([]*megaport.MCR, error) {
	return nil, nil
}
func (m *MockMCRService) GetMCR(ctx context.Context, mcrId string) (*megaport.MCR, error) {
	return nil, nil
}
func (m *MockMCRService) CreatePrefixFilterList(ctx context.Context, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	return nil, nil
}
func (m *MockMCRService) ListMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	return nil, nil
}
func (m *MockMCRService) GetMCRPrefixFilterList(ctx context.Context, mcrID string, id int) (*megaport.MCRPrefixFilterList, error) {
	return nil, nil
}
func (m *MockMCRService) ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, id int, list *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	return nil, nil
}
func (m *MockMCRService) DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, id int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	return nil, nil
}
func (m *MockMCRService) ModifyMCR(ctx context.Context, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	return nil, nil
}
func (m *MockMCRService) DeleteMCR(ctx context.Context, req *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	return nil, nil
}
func (m *MockMCRService) RestoreMCR(ctx context.Context, mcrId string) (*megaport.RestoreMCRResponse, error) {
	return nil, nil
}
func (m *MockMCRService) ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	return nil, nil
}
func (m *MockMCRService) UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	return nil
}
func (m *MockMCRService) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	return nil, nil
}

// MockMVEService implements megaport.MVEService for testing.
type MockMVEService struct {
	BuyMVEResult        *megaport.BuyMVEResponse
	BuyMVEErr           error
	ValidateMVEOrderErr error
}

func (m *MockMVEService) BuyMVE(ctx context.Context, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	if m.BuyMVEErr != nil {
		return nil, m.BuyMVEErr
	}
	if m.BuyMVEResult != nil {
		return m.BuyMVEResult, nil
	}
	return &megaport.BuyMVEResponse{TechnicalServiceUID: "mve-uid-mock"}, nil
}

func (m *MockMVEService) ValidateMVEOrder(ctx context.Context, req *megaport.BuyMVERequest) error {
	return m.ValidateMVEOrderErr
}

func (m *MockMVEService) ListMVEs(ctx context.Context, req *megaport.ListMVEsRequest) ([]*megaport.MVE, error) {
	return nil, nil
}
func (m *MockMVEService) GetMVE(ctx context.Context, mveId string) (*megaport.MVE, error) {
	return nil, nil
}
func (m *MockMVEService) ModifyMVE(ctx context.Context, req *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	return nil, nil
}
func (m *MockMVEService) DeleteMVE(ctx context.Context, req *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	return nil, nil
}
func (m *MockMVEService) ListMVEImages(ctx context.Context) ([]*megaport.MVEImage, error) {
	return nil, nil
}
func (m *MockMVEService) ListAvailableMVESizes(ctx context.Context) ([]*megaport.MVESize, error) {
	return nil, nil
}
func (m *MockMVEService) ListMVEResourceTags(ctx context.Context, mveID string) (map[string]string, error) {
	return nil, nil
}
func (m *MockMVEService) UpdateMVEResourceTags(ctx context.Context, mveID string, tags map[string]string) error {
	return nil
}

// MockVXCService implements megaport.VXCService for testing.
type MockVXCService struct {
	BuyVXCResult        *megaport.BuyVXCResponse
	BuyVXCErr           error
	ValidateVXCOrderErr error
	CapturedVXCRequest  *megaport.BuyVXCRequest
}

func (m *MockVXCService) BuyVXC(ctx context.Context, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	m.CapturedVXCRequest = req
	if m.BuyVXCErr != nil {
		return nil, m.BuyVXCErr
	}
	if m.BuyVXCResult != nil {
		return m.BuyVXCResult, nil
	}
	return &megaport.BuyVXCResponse{TechnicalServiceUID: "vxc-uid-mock"}, nil
}

func (m *MockVXCService) ValidateVXCOrder(ctx context.Context, req *megaport.BuyVXCRequest) error {
	return m.ValidateVXCOrderErr
}

func (m *MockVXCService) ListVXCs(ctx context.Context, req *megaport.ListVXCsRequest) ([]*megaport.VXC, error) {
	return nil, nil
}
func (m *MockVXCService) GetVXC(ctx context.Context, id string) (*megaport.VXC, error) {
	return nil, nil
}
func (m *MockVXCService) DeleteVXC(ctx context.Context, id string, req *megaport.DeleteVXCRequest) error {
	return nil
}
func (m *MockVXCService) UpdateVXC(ctx context.Context, id string, req *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	return nil, nil
}
func (m *MockVXCService) LookupPartnerPorts(ctx context.Context, req *megaport.LookupPartnerPortsRequest) (*megaport.LookupPartnerPortsResponse, error) {
	return nil, nil
}
func (m *MockVXCService) ListPartnerPorts(ctx context.Context, req *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return nil, nil
}
func (m *MockVXCService) ListVXCResourceTags(ctx context.Context, vxcID string) (map[string]string, error) {
	return nil, nil
}
func (m *MockVXCService) UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error {
	return nil
}
