package mve

import (
	"context"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/mock"
)

type MockMVEService struct {
	mock.Mock
	BuyMVEError                          error
	BuyMVEResult                         *megaport.BuyMVEResponse
	CapturedBuyMVERequest                *megaport.BuyMVERequest
	ValidateMVEOrderError                error
	GetMVEError                          error
	GetMVEResult                         *megaport.MVE
	ModifyMVEError                       error
	ModifyMVEResult                      *megaport.ModifyMVEResponse
	CapturedModifyMVERequest             *megaport.ModifyMVERequest
	DeleteMVEError                       error
	DeleteMVEResult                      *megaport.DeleteMVEResponse
	CapturedDeleteMVERequest             *megaport.DeleteMVERequest
	ListMVEImagesError                   error
	ListMVEImagesResult                  []*megaport.MVEImage
	ListAvailableMVESizesError           error
	ListAvailableMVESizesResult          []*megaport.MVESize
	ListMVEResourceTagsError             error
	ListMVEResourceTagsResult            map[string]string
	CapturedListMVEResourceTagsUID       string
	UpdateMVEResourceTagsError           error
	CapturedUpdateMVEResourceTagsUID     string
	CapturedUpdateMVEResourceTagsTags    map[string]string
	ListMVEsError                        error
	ListMVEsResult                       []*megaport.MVE
	CapturedListMVEsRequest              *megaport.ListMVEsRequest
	RestoreMVEError                      error
	ListMVEResourceTagsErr               error
	UpdateMVEResourceTagsErr             error
	CapturedUpdateMVEResourceTagsRequest map[string]string
}

func (m *MockMVEService) Reset() {
	m.ValidateMVEOrderError = nil
	m.BuyMVEResult = nil
	m.BuyMVEError = nil
	m.ModifyMVEResult = nil
	m.ModifyMVEError = nil
	m.CapturedBuyMVERequest = nil
	m.CapturedModifyMVERequest = nil
	m.GetMVEResult = nil
	m.GetMVEError = nil
	m.ListMVEsResult = nil
	m.ListMVEsError = nil
	m.DeleteMVEResult = nil
	m.DeleteMVEError = nil
	m.RestoreMVEError = nil
	m.ListMVEImagesResult = nil
	m.ListMVEImagesError = nil
	m.ListAvailableMVESizesResult = nil
	m.ListAvailableMVESizesError = nil
	m.CapturedListMVEsRequest = nil
	m.ListMVEResourceTagsResult = nil
	m.ListMVEResourceTagsErr = nil
	m.UpdateMVEResourceTagsErr = nil
	m.CapturedUpdateMVEResourceTagsRequest = nil
}

func (m *MockMVEService) BuyMVE(ctx context.Context, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	m.CapturedBuyMVERequest = req
	if m.BuyMVEError != nil {
		return nil, m.BuyMVEError
	}
	if m.BuyMVEResult != nil {
		return m.BuyMVEResult, nil
	}
	return &megaport.BuyMVEResponse{TechnicalServiceUID: "default-mve-uid"}, nil
}

func (m *MockMVEService) ValidateMVEOrder(ctx context.Context, req *megaport.BuyMVERequest) error {
	if m.ValidateMVEOrderError != nil {
		return m.ValidateMVEOrderError
	}
	return nil
}

func (m *MockMVEService) ListMVEs(ctx context.Context, req *megaport.ListMVEsRequest) ([]*megaport.MVE, error) {
	m.CapturedListMVEsRequest = req
	if m.ListMVEsError != nil {
		return nil, m.ListMVEsError
	}
	if m.ListMVEsResult != nil {
		return m.ListMVEsResult, nil
	}
	return []*megaport.MVE{}, nil
}

func (m *MockMVEService) GetMVE(ctx context.Context, mveId string) (*megaport.MVE, error) {
	if m.GetMVEError != nil {
		return nil, m.GetMVEError
	}
	if m.GetMVEResult != nil {
		return m.GetMVEResult, nil
	}
	return &megaport.MVE{
		UID:                mveId,
		Name:               "Mock MVE",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockMVEService) ModifyMVE(ctx context.Context, req *megaport.ModifyMVERequest) (*megaport.ModifyMVEResponse, error) {
	m.CapturedModifyMVERequest = req
	if m.ModifyMVEError != nil {
		return nil, m.ModifyMVEError
	}
	if m.ModifyMVEResult != nil {
		return m.ModifyMVEResult, nil
	}
	return &megaport.ModifyMVEResponse{
		MVEUpdated: true,
	}, nil
}

func (m *MockMVEService) DeleteMVE(ctx context.Context, req *megaport.DeleteMVERequest) (*megaport.DeleteMVEResponse, error) {
	m.CapturedDeleteMVERequest = req
	if m.DeleteMVEError != nil {
		return nil, m.DeleteMVEError
	}
	if m.DeleteMVEResult != nil {
		return m.DeleteMVEResult, nil
	}
	return &megaport.DeleteMVEResponse{
		IsDeleted: true,
	}, nil
}

func (m *MockMVEService) ListMVEImages(ctx context.Context) ([]*megaport.MVEImage, error) {
	if m.ListMVEImagesError != nil {
		return nil, m.ListMVEImagesError
	}
	if m.ListMVEImagesResult != nil {
		return m.ListMVEImagesResult, nil
	}
	return []*megaport.MVEImage{}, nil
}

func (m *MockMVEService) ListAvailableMVESizes(ctx context.Context) ([]*megaport.MVESize, error) {
	if m.ListAvailableMVESizesError != nil {
		return nil, m.ListAvailableMVESizesError
	}
	if m.ListAvailableMVESizesResult != nil {
		return m.ListAvailableMVESizesResult, nil
	}
	return []*megaport.MVESize{}, nil
}

func (m *MockMVEService) ListMVEResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	if m.ListMVEResourceTagsErr != nil {
		return nil, m.ListMVEResourceTagsErr
	}
	return m.ListMVEResourceTagsResult, nil
}

func (m *MockMVEService) UpdateMVEResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	m.CapturedUpdateMVEResourceTagsRequest = tags
	return m.UpdateMVEResourceTagsErr
}
