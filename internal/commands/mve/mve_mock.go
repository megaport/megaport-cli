package mve

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// MockMVEService implements the required MVE service methods for testing
type MockMVEService struct {
	// Optional fields to customize behavior
	GetMVEErr           error
	GetMVEResult        *megaport.MVE
	ListMVEsErr         error
	ListMVEsResult      []*megaport.MVE
	BuyMVEErr           error
	BuyMVEResult        *megaport.BuyMVEResponse
	DeleteMVEErr        error
	DeleteMVEResult     *megaport.DeleteMVEResponse
	ModifyMVEErr        error
	ModifyMVEResult     *megaport.ModifyMVEResponse
	ValidateMVEOrderErr error

	// Resource tags related fields
	ListMVEResourceTagsErr    error
	ListMVEResourceTagsResult map[string]string
	UpdateMVEResourceTagsErr  error

	// Images and sizes
	ListMVEImagesErr            error
	ListMVEImagesResult         []*megaport.MVEImage
	ListAvailableMVESizesErr    error
	ListAvailableMVESizesResult []*megaport.MVESize

	// For tracking request parameters in tests
	CapturedBuyMVERequest                *megaport.BuyMVERequest
	CapturedModifyMVERequest             *megaport.ModifyMVERequest
	CapturedListMVEsRequest              *megaport.ListMVEsRequest
	CapturedUpdateMVEResourceTagsRequest map[string]string
}

func (m *MockMVEService) GetMVE(ctx context.Context, mveID string) (*megaport.MVE, error) {
	if m.GetMVEErr != nil {
		return nil, m.GetMVEErr
	}
	if m.GetMVEResult != nil {
		return m.GetMVEResult, nil
	}
	return &megaport.MVE{
		UID:                mveID,
		Name:               "Mock MVE",
		ProvisioningStatus: "LIVE",
		Vendor:             "cisco",
		Size:               "MEDIUM",
	}, nil
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

func (m *MockMVEService) BuyMVE(ctx context.Context, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	m.CapturedBuyMVERequest = req
	if m.BuyMVEErr != nil {
		return nil, m.BuyMVEErr
	}
	if m.BuyMVEResult != nil {
		return m.BuyMVEResult, nil
	}
	return &megaport.BuyMVEResponse{
		TechnicalServiceUID: "mock-mve-uid",
	}, nil
}

func (m *MockMVEService) ValidateMVEOrder(ctx context.Context, req *megaport.BuyMVERequest) error {
	return m.ValidateMVEOrderErr
}

func (m *MockMVEService) DeleteMVE(ctx context.Context, req *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	if m.DeleteMVEErr != nil {
		return nil, m.DeleteMVEErr
	}
	if m.DeleteMVEResult != nil {
		return m.DeleteMVEResult, nil
	}
	return &megaport.DeleteMVEResponse{
		IsDeleted: true,
	}, nil
}

func (m *MockMVEService) ModifyMVE(ctx context.Context, req *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	m.CapturedModifyMVERequest = req
	if m.ModifyMVEErr != nil {
		return nil, m.ModifyMVEErr
	}
	if m.ModifyMVEResult != nil {
		return m.ModifyMVEResult, nil
	}
	return &megaport.ModifyMVEResponse{
		MVEUpdated: true,
	}, nil
}

func (m *MockMVEService) ListMVEResourceTags(ctx context.Context, mveID string) (map[string]string, error) {
	if m.ListMVEResourceTagsErr != nil {
		return nil, m.ListMVEResourceTagsErr
	}
	if m.ListMVEResourceTagsResult != nil {
		return m.ListMVEResourceTagsResult, nil
	}
	return map[string]string{
		"environment": "test",
		"owner":       "automation",
	}, nil
}

func (m *MockMVEService) UpdateMVEResourceTags(ctx context.Context, mveID string, tags map[string]string) error {
	m.CapturedUpdateMVEResourceTagsRequest = tags
	return m.UpdateMVEResourceTagsErr
}

func (m *MockMVEService) ListMVEImages(ctx context.Context) ([]*megaport.MVEImage, error) {
	if m.ListMVEImagesErr != nil {
		return nil, m.ListMVEImagesErr
	}
	if m.ListMVEImagesResult != nil {
		return m.ListMVEImagesResult, nil
	}
	return []*megaport.MVEImage{}, nil
}

func (m *MockMVEService) ListAvailableMVESizes(ctx context.Context) ([]*megaport.MVESize, error) {
	if m.ListAvailableMVESizesErr != nil {
		return nil, m.ListAvailableMVESizesErr
	}
	if m.ListAvailableMVESizesResult != nil {
		return m.ListAvailableMVESizesResult, nil
	}
	return []*megaport.MVESize{}, nil
}

// Reset clears all captured values and resets errors to nil
func (m *MockMVEService) Reset() {
	// Reset errors
	m.GetMVEErr = nil
	m.ListMVEsErr = nil
	m.BuyMVEErr = nil
	m.DeleteMVEErr = nil
	m.ModifyMVEErr = nil
	m.ValidateMVEOrderErr = nil
	m.ListMVEResourceTagsErr = nil
	m.UpdateMVEResourceTagsErr = nil
	m.ListMVEImagesErr = nil
	m.ListAvailableMVESizesErr = nil

	// Reset captured requests
	m.CapturedBuyMVERequest = nil
	m.CapturedModifyMVERequest = nil
	m.CapturedListMVEsRequest = nil
	m.CapturedUpdateMVEResourceTagsRequest = nil
}
