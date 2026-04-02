package status

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// MockPortService is a minimal mock for testing the status dashboard.
type MockPortService struct {
	ListPortsResult []*megaport.Port
	ListPortsErr    error
}

func (m *MockPortService) ListPorts(ctx context.Context) ([]*megaport.Port, error) {
	if m.ListPortsErr != nil {
		return nil, m.ListPortsErr
	}
	if m.ListPortsResult != nil {
		return m.ListPortsResult, nil
	}
	return []*megaport.Port{}, nil
}

func (m *MockPortService) GetPort(_ context.Context, _ string) (*megaport.Port, error) {
	return nil, nil
}

func (m *MockPortService) BuyPort(_ context.Context, _ *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return nil, nil
}

func (m *MockPortService) CheckPortVLANAvailability(_ context.Context, _ string, _ int) (bool, error) {
	return false, nil
}

func (m *MockPortService) DeletePort(_ context.Context, _ *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return nil, nil
}

func (m *MockPortService) ListPortResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, nil
}

func (m *MockPortService) ValidatePortOrder(_ context.Context, _ *megaport.BuyPortRequest) error {
	return nil
}

func (m *MockPortService) LockPort(_ context.Context, _ string) (*megaport.LockPortResponse, error) {
	return nil, nil
}

func (m *MockPortService) ModifyPort(_ context.Context, _ *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return nil, nil
}

func (m *MockPortService) RestorePort(_ context.Context, _ string) (*megaport.RestorePortResponse, error) {
	return nil, nil
}

func (m *MockPortService) UnlockPort(_ context.Context, _ string) (*megaport.UnlockPortResponse, error) {
	return nil, nil
}

func (m *MockPortService) UpdatePortResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return nil
}

// MockMCRService is a minimal mock for testing the status dashboard.
type MockMCRService struct {
	ListMCRsResult          []*megaport.MCR
	ListMCRsErr             error
	CapturedListMCRsRequest *megaport.ListMCRsRequest
}

func (m *MockMCRService) ListMCRs(ctx context.Context, req *megaport.ListMCRsRequest) ([]*megaport.MCR, error) {
	m.CapturedListMCRsRequest = req
	if m.ListMCRsErr != nil {
		return nil, m.ListMCRsErr
	}
	if m.ListMCRsResult != nil {
		return m.ListMCRsResult, nil
	}
	return []*megaport.MCR{}, nil
}

func (m *MockMCRService) BuyMCR(_ context.Context, _ *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	return nil, nil
}

func (m *MockMCRService) ValidateMCROrder(_ context.Context, _ *megaport.BuyMCRRequest) error {
	return nil
}

func (m *MockMCRService) GetMCR(_ context.Context, _ string) (*megaport.MCR, error) {
	return nil, nil
}

func (m *MockMCRService) DeleteMCR(_ context.Context, _ *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	return nil, nil
}

func (m *MockMCRService) RestoreMCR(_ context.Context, _ string) (*megaport.RestoreMCRResponse, error) {
	return nil, nil
}

func (m *MockMCRService) CreatePrefixFilterList(_ context.Context, _ *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	return nil, nil
}

func (m *MockMCRService) ListMCRPrefixFilterLists(_ context.Context, _ string) ([]*megaport.PrefixFilterList, error) {
	return nil, nil
}

func (m *MockMCRService) GetMCRPrefixFilterList(_ context.Context, _ string, _ int) (*megaport.MCRPrefixFilterList, error) {
	return nil, nil
}

func (m *MockMCRService) ModifyMCRPrefixFilterList(_ context.Context, _ string, _ int, _ *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	return nil, nil
}

func (m *MockMCRService) DeleteMCRPrefixFilterList(_ context.Context, _ string, _ int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	return nil, nil
}

func (m *MockMCRService) ModifyMCR(_ context.Context, _ *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	return nil, nil
}

func (m *MockMCRService) ListMCRResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, nil
}

func (m *MockMCRService) UpdateMCRResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return nil
}

func (m *MockMCRService) GetMCRPrefixFilterLists(_ context.Context, _ string) ([]*megaport.PrefixFilterList, error) {
	return nil, nil
}

// MockMVEService is a minimal mock for testing the status dashboard.
type MockMVEService struct {
	ListMVEsResult          []*megaport.MVE
	ListMVEsErr             error
	CapturedListMVEsRequest *megaport.ListMVEsRequest
}

func (m *MockMVEService) ListMVEs(ctx context.Context, req *megaport.ListMVEsRequest) ([]*megaport.MVE, error) {
	m.CapturedListMVEsRequest = req
	if m.ListMVEsErr != nil {
		return nil, m.ListMVEsErr
	}
	if m.ListMVEsResult != nil {
		return m.ListMVEsResult, nil
	}
	return []*megaport.MVE{}, nil
}

func (m *MockMVEService) GetMVE(_ context.Context, _ string) (*megaport.MVE, error) {
	return nil, nil
}

func (m *MockMVEService) BuyMVE(_ context.Context, _ *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	return nil, nil
}

func (m *MockMVEService) ValidateMVEOrder(_ context.Context, _ *megaport.BuyMVERequest) error {
	return nil
}

func (m *MockMVEService) DeleteMVE(_ context.Context, _ *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	return nil, nil
}

func (m *MockMVEService) ModifyMVE(_ context.Context, _ *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	return nil, nil
}

func (m *MockMVEService) ListMVEResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, nil
}

func (m *MockMVEService) UpdateMVEResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return nil
}

func (m *MockMVEService) ListMVEImages(_ context.Context) ([]*megaport.MVEImage, error) {
	return nil, nil
}

func (m *MockMVEService) ListAvailableMVESizes(_ context.Context) ([]*megaport.MVESize, error) {
	return nil, nil
}

// MockVXCService is a minimal mock for testing the status dashboard.
type MockVXCService struct {
	ListVXCsResult          []*megaport.VXC
	ListVXCsErr             error
	CapturedListVXCsRequest *megaport.ListVXCsRequest
}

func (m *MockVXCService) ListVXCs(ctx context.Context, req *megaport.ListVXCsRequest) ([]*megaport.VXC, error) {
	m.CapturedListVXCsRequest = req
	if m.ListVXCsErr != nil {
		return nil, m.ListVXCsErr
	}
	if m.ListVXCsResult != nil {
		return m.ListVXCsResult, nil
	}
	return []*megaport.VXC{}, nil
}

func (m *MockVXCService) BuyVXC(_ context.Context, _ *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	return nil, nil
}

func (m *MockVXCService) ValidateVXCOrder(_ context.Context, _ *megaport.BuyVXCRequest) error {
	return nil
}

func (m *MockVXCService) GetVXC(_ context.Context, _ string) (*megaport.VXC, error) {
	return nil, nil
}

func (m *MockVXCService) DeleteVXC(_ context.Context, _ string, _ *megaport.DeleteVXCRequest) error {
	return nil
}

func (m *MockVXCService) UpdateVXC(_ context.Context, _ string, _ *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	return nil, nil
}

func (m *MockVXCService) LookupPartnerPorts(_ context.Context, _ *megaport.LookupPartnerPortsRequest) (*megaport.LookupPartnerPortsResponse, error) {
	return nil, nil
}

func (m *MockVXCService) ListPartnerPorts(_ context.Context, _ *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return nil, nil
}

func (m *MockVXCService) ListVXCResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, nil
}

func (m *MockVXCService) UpdateVXCResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return nil
}

// MockIXService is a minimal mock for testing the status dashboard.
type MockIXService struct {
	ListIXsResult          []*megaport.IX
	ListIXsErr             error
	CapturedListIXsRequest *megaport.ListIXsRequest
}

func (m *MockIXService) ListIXs(ctx context.Context, req *megaport.ListIXsRequest) ([]*megaport.IX, error) {
	m.CapturedListIXsRequest = req
	if m.ListIXsErr != nil {
		return nil, m.ListIXsErr
	}
	if m.ListIXsResult != nil {
		return m.ListIXsResult, nil
	}
	return []*megaport.IX{}, nil
}

func (m *MockIXService) BuyIX(_ context.Context, _ *megaport.BuyIXRequest) (*megaport.BuyIXResponse, error) {
	return nil, nil
}

func (m *MockIXService) ValidateIXOrder(_ context.Context, _ *megaport.BuyIXRequest) error {
	return nil
}

func (m *MockIXService) GetIX(_ context.Context, _ string) (*megaport.IX, error) {
	return nil, nil
}

func (m *MockIXService) UpdateIX(_ context.Context, _ string, _ *megaport.UpdateIXRequest) (*megaport.IX, error) {
	return nil, nil
}

func (m *MockIXService) DeleteIX(_ context.Context, _ string, _ *megaport.DeleteIXRequest) error {
	return nil
}
