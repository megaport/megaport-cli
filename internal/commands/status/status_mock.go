package status

import (
	"context"
	"fmt"

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
	return nil, fmt.Errorf("mock: GetPort not configured")
}

func (m *MockPortService) BuyPort(_ context.Context, _ *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return nil, fmt.Errorf("mock: BuyPort not configured")
}

func (m *MockPortService) CheckPortVLANAvailability(_ context.Context, _ string, _ int) (bool, error) {
	return false, fmt.Errorf("mock: CheckPortVLANAvailability not configured")
}

func (m *MockPortService) DeletePort(_ context.Context, _ *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return nil, fmt.Errorf("mock: DeletePort not configured")
}

func (m *MockPortService) ListPortResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListPortResourceTags not configured")
}

func (m *MockPortService) ValidatePortOrder(_ context.Context, _ *megaport.BuyPortRequest) error {
	return fmt.Errorf("mock: ValidatePortOrder not configured")
}

func (m *MockPortService) LockPort(_ context.Context, _ string) (*megaport.LockPortResponse, error) {
	return nil, fmt.Errorf("mock: LockPort not configured")
}

func (m *MockPortService) ModifyPort(_ context.Context, _ *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return nil, fmt.Errorf("mock: ModifyPort not configured")
}

func (m *MockPortService) RestorePort(_ context.Context, _ string) (*megaport.RestorePortResponse, error) {
	return nil, fmt.Errorf("mock: RestorePort not configured")
}

func (m *MockPortService) UnlockPort(_ context.Context, _ string) (*megaport.UnlockPortResponse, error) {
	return nil, fmt.Errorf("mock: UnlockPort not configured")
}

func (m *MockPortService) UpdatePortResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return fmt.Errorf("mock: UpdatePortResourceTags not configured")
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
	return nil, fmt.Errorf("mock: BuyMCR not configured")
}

func (m *MockMCRService) ValidateMCROrder(_ context.Context, _ *megaport.BuyMCRRequest) error {
	return fmt.Errorf("mock: ValidateMCROrder not configured")
}

func (m *MockMCRService) GetMCR(_ context.Context, _ string) (*megaport.MCR, error) {
	return nil, fmt.Errorf("mock: GetMCR not configured")
}

func (m *MockMCRService) DeleteMCR(_ context.Context, _ *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMCR not configured")
}

func (m *MockMCRService) RestoreMCR(_ context.Context, _ string) (*megaport.RestoreMCRResponse, error) {
	return nil, fmt.Errorf("mock: RestoreMCR not configured")
}

func (m *MockMCRService) CreatePrefixFilterList(_ context.Context, _ *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: CreatePrefixFilterList not configured")
}

func (m *MockMCRService) ListMCRPrefixFilterLists(_ context.Context, _ string) ([]*megaport.PrefixFilterList, error) {
	return nil, fmt.Errorf("mock: ListMCRPrefixFilterLists not configured")
}

func (m *MockMCRService) GetMCRPrefixFilterList(_ context.Context, _ string, _ int) (*megaport.MCRPrefixFilterList, error) {
	return nil, fmt.Errorf("mock: GetMCRPrefixFilterList not configured")
}

func (m *MockMCRService) ModifyMCRPrefixFilterList(_ context.Context, _ string, _ int, _ *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMCRPrefixFilterList not configured")
}

func (m *MockMCRService) DeleteMCRPrefixFilterList(_ context.Context, _ string, _ int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMCRPrefixFilterList not configured")
}

func (m *MockMCRService) ModifyMCR(_ context.Context, _ *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMCR not configured")
}

func (m *MockMCRService) ListMCRResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListMCRResourceTags not configured")
}

func (m *MockMCRService) UpdateMCRResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return fmt.Errorf("mock: UpdateMCRResourceTags not configured")
}

func (m *MockMCRService) GetMCRPrefixFilterLists(_ context.Context, _ string) ([]*megaport.PrefixFilterList, error) {
	return nil, fmt.Errorf("mock: GetMCRPrefixFilterLists not configured")
}

func (m *MockMCRService) UpdateMCRWithAddOn(_ context.Context, _ string, _ megaport.MCRAddOnRequest) error {
	return fmt.Errorf("mock: UpdateMCRWithAddOn not configured")
}

func (m *MockMCRService) UpdateMCRIPsecAddOn(_ context.Context, _ string, _ string, _ int) error {
	return fmt.Errorf("mock: UpdateMCRIPsecAddOn not configured")
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
	return nil, fmt.Errorf("mock: GetMVE not configured")
}

func (m *MockMVEService) BuyMVE(_ context.Context, _ *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	return nil, fmt.Errorf("mock: BuyMVE not configured")
}

func (m *MockMVEService) ValidateMVEOrder(_ context.Context, _ *megaport.BuyMVERequest) error {
	return fmt.Errorf("mock: ValidateMVEOrder not configured")
}

func (m *MockMVEService) DeleteMVE(_ context.Context, _ *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMVE not configured")
}

func (m *MockMVEService) ModifyMVE(_ context.Context, _ *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMVE not configured")
}

func (m *MockMVEService) ListMVEResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListMVEResourceTags not configured")
}

func (m *MockMVEService) UpdateMVEResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return fmt.Errorf("mock: UpdateMVEResourceTags not configured")
}

func (m *MockMVEService) ListMVEImages(_ context.Context) ([]*megaport.MVEImage, error) {
	return nil, fmt.Errorf("mock: ListMVEImages not configured")
}

func (m *MockMVEService) ListAvailableMVESizes(_ context.Context) ([]*megaport.MVESize, error) {
	return nil, fmt.Errorf("mock: ListAvailableMVESizes not configured")
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
	return nil, fmt.Errorf("mock: BuyVXC not configured")
}

func (m *MockVXCService) ValidateVXCOrder(_ context.Context, _ *megaport.BuyVXCRequest) error {
	return fmt.Errorf("mock: ValidateVXCOrder not configured")
}

func (m *MockVXCService) GetVXC(_ context.Context, _ string) (*megaport.VXC, error) {
	return nil, fmt.Errorf("mock: GetVXC not configured")
}

func (m *MockVXCService) DeleteVXC(_ context.Context, _ string, _ *megaport.DeleteVXCRequest) error {
	return fmt.Errorf("mock: DeleteVXC not configured")
}

func (m *MockVXCService) UpdateVXC(_ context.Context, _ string, _ *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	return nil, fmt.Errorf("mock: UpdateVXC not configured")
}

func (m *MockVXCService) LookupPartnerPorts(_ context.Context, _ *megaport.LookupPartnerPortsRequest) (*megaport.LookupPartnerPortsResponse, error) {
	return nil, fmt.Errorf("mock: LookupPartnerPorts not configured")
}

func (m *MockVXCService) ListPartnerPorts(_ context.Context, _ *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return nil, fmt.Errorf("mock: ListPartnerPorts not configured")
}

func (m *MockVXCService) ListVXCResourceTags(_ context.Context, _ string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListVXCResourceTags not configured")
}

func (m *MockVXCService) UpdateVXCResourceTags(_ context.Context, _ string, _ map[string]string) error {
	return fmt.Errorf("mock: UpdateVXCResourceTags not configured")
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
	return nil, fmt.Errorf("mock: BuyIX not configured")
}

func (m *MockIXService) ValidateIXOrder(_ context.Context, _ *megaport.BuyIXRequest) error {
	return fmt.Errorf("mock: ValidateIXOrder not configured")
}

func (m *MockIXService) GetIX(_ context.Context, _ string) (*megaport.IX, error) {
	return nil, fmt.Errorf("mock: GetIX not configured")
}

func (m *MockIXService) UpdateIX(_ context.Context, _ string, _ *megaport.UpdateIXRequest) (*megaport.IX, error) {
	return nil, fmt.Errorf("mock: UpdateIX not configured")
}

func (m *MockIXService) DeleteIX(_ context.Context, _ string, _ *megaport.DeleteIXRequest) error {
	return fmt.Errorf("mock: DeleteIX not configured")
}
