package topology

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
)

// MockPortService satisfies megaport.PortService for testing.
type MockPortService struct {
	ListPortsResult []*megaport.Port
	ListPortsErr    error
}

func (m *MockPortService) ListPorts(ctx context.Context) ([]*megaport.Port, error) {
	return m.ListPortsResult, m.ListPortsErr
}

func (m *MockPortService) BuyPort(ctx context.Context, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return nil, fmt.Errorf("mock: BuyPort not configured")
}
func (m *MockPortService) ValidatePortOrder(ctx context.Context, req *megaport.BuyPortRequest) error {
	return fmt.Errorf("mock: ValidatePortOrder not configured")
}
func (m *MockPortService) GetPort(ctx context.Context, portId string) (*megaport.Port, error) {
	return nil, fmt.Errorf("mock: GetPort not configured")
}
func (m *MockPortService) ModifyPort(ctx context.Context, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return nil, fmt.Errorf("mock: ModifyPort not configured")
}
func (m *MockPortService) DeletePort(ctx context.Context, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return nil, fmt.Errorf("mock: DeletePort not configured")
}
func (m *MockPortService) RestorePort(ctx context.Context, portId string) (*megaport.RestorePortResponse, error) {
	return nil, fmt.Errorf("mock: RestorePort not configured")
}
func (m *MockPortService) LockPort(ctx context.Context, portId string) (*megaport.LockPortResponse, error) {
	return nil, fmt.Errorf("mock: LockPort not configured")
}
func (m *MockPortService) UnlockPort(ctx context.Context, portId string) (*megaport.UnlockPortResponse, error) {
	return nil, fmt.Errorf("mock: UnlockPort not configured")
}
func (m *MockPortService) CheckPortVLANAvailability(ctx context.Context, portId string, vlan int) (bool, error) {
	return false, fmt.Errorf("mock: CheckPortVLANAvailability not configured")
}
func (m *MockPortService) ListPortResourceTags(ctx context.Context, portID string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListPortResourceTags not configured")
}
func (m *MockPortService) UpdatePortResourceTags(ctx context.Context, portID string, tags map[string]string) error {
	return fmt.Errorf("mock: UpdatePortResourceTags not configured")
}

// MockMCRService satisfies megaport.MCRService for testing.
type MockMCRService struct {
	ListMCRsResult []*megaport.MCR
	ListMCRsErr    error
}

func (m *MockMCRService) ListMCRs(ctx context.Context, req *megaport.ListMCRsRequest) ([]*megaport.MCR, error) {
	return m.ListMCRsResult, m.ListMCRsErr
}

func (m *MockMCRService) BuyMCR(ctx context.Context, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	return nil, fmt.Errorf("mock: BuyMCR not configured")
}
func (m *MockMCRService) ValidateMCROrder(ctx context.Context, req *megaport.BuyMCRRequest) error {
	return fmt.Errorf("mock: ValidateMCROrder not configured")
}
func (m *MockMCRService) GetMCR(ctx context.Context, mcrId string) (*megaport.MCR, error) {
	return nil, fmt.Errorf("mock: GetMCR not configured")
}
func (m *MockMCRService) CreatePrefixFilterList(ctx context.Context, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: CreatePrefixFilterList not configured")
}
func (m *MockMCRService) ListMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	return nil, fmt.Errorf("mock: ListMCRPrefixFilterLists not configured")
}
func (m *MockMCRService) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	return nil, fmt.Errorf("mock: GetMCRPrefixFilterLists not configured")
}
func (m *MockMCRService) GetMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	return nil, fmt.Errorf("mock: GetMCRPrefixFilterList not configured")
}
func (m *MockMCRService) ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMCRPrefixFilterList not configured")
}
func (m *MockMCRService) DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMCRPrefixFilterList not configured")
}
func (m *MockMCRService) ModifyMCR(ctx context.Context, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMCR not configured")
}
func (m *MockMCRService) DeleteMCR(ctx context.Context, req *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMCR not configured")
}
func (m *MockMCRService) RestoreMCR(ctx context.Context, mcrId string) (*megaport.RestoreMCRResponse, error) {
	return nil, fmt.Errorf("mock: RestoreMCR not configured")
}
func (m *MockMCRService) ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListMCRResourceTags not configured")
}
func (m *MockMCRService) UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	return fmt.Errorf("mock: UpdateMCRResourceTags not configured")
}

func (m *MockMCRService) UpdateMCRWithAddOn(ctx context.Context, mcrID string, req megaport.MCRAddOnRequest) error {
	return fmt.Errorf("mock: UpdateMCRWithAddOn not configured")
}

func (m *MockMCRService) UpdateMCRIPsecAddOn(ctx context.Context, mcrID string, addOnUID string, tunnelCount int) error {
	return fmt.Errorf("mock: UpdateMCRIPsecAddOn not configured")
}

func (m *MockMCRService) WaitForMCRReady(ctx context.Context, mcrID string, timeout time.Duration) error {
	return fmt.Errorf("mock: WaitForMCRReady not configured")
}

// MockMVEService satisfies megaport.MVEService for testing.
type MockMVEService struct {
	ListMVEsResult []*megaport.MVE
	ListMVEsErr    error
}

func (m *MockMVEService) ListMVEs(ctx context.Context, req *megaport.ListMVEsRequest) ([]*megaport.MVE, error) {
	return m.ListMVEsResult, m.ListMVEsErr
}

func (m *MockMVEService) BuyMVE(ctx context.Context, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	return nil, fmt.Errorf("mock: BuyMVE not configured")
}
func (m *MockMVEService) ValidateMVEOrder(ctx context.Context, req *megaport.BuyMVERequest) error {
	return fmt.Errorf("mock: ValidateMVEOrder not configured")
}
func (m *MockMVEService) GetMVE(ctx context.Context, mveId string) (*megaport.MVE, error) {
	return nil, fmt.Errorf("mock: GetMVE not configured")
}
func (m *MockMVEService) ModifyMVE(ctx context.Context, req *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	return nil, fmt.Errorf("mock: ModifyMVE not configured")
}
func (m *MockMVEService) DeleteMVE(ctx context.Context, req *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	return nil, fmt.Errorf("mock: DeleteMVE not configured")
}
func (m *MockMVEService) ListMVEImages(ctx context.Context) ([]*megaport.MVEImage, error) {
	return nil, fmt.Errorf("mock: ListMVEImages not configured")
}
func (m *MockMVEService) ListAvailableMVESizes(ctx context.Context) ([]*megaport.MVESize, error) {
	return nil, fmt.Errorf("mock: ListAvailableMVESizes not configured")
}
func (m *MockMVEService) ListMVEResourceTags(ctx context.Context, mveID string) (map[string]string, error) {
	return nil, fmt.Errorf("mock: ListMVEResourceTags not configured")
}
func (m *MockMVEService) UpdateMVEResourceTags(ctx context.Context, mveID string, tags map[string]string) error {
	return fmt.Errorf("mock: UpdateMVEResourceTags not configured")
}
