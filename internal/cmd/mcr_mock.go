package cmd

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockMCRService struct {
	GetMCRErr                          error
	GetMCRResult                       *megaport.MCR
	ListMCRsErr                        error
	ListMCRsResult                     []*megaport.MCR
	BuyMCRErr                          error
	BuyMCRResult                       *megaport.BuyMCRResponse
	CapturedBuyMCRRequest              *megaport.BuyMCRRequest
	ValidateMCROrderErr                error
	ModifyMCRErr                       error
	ModifyMCRResult                    *megaport.ModifyMCRResponse
	CapturedModifyMCRRequest           *megaport.ModifyMCRRequest
	DeleteMCRErr                       error
	DeleteMCRResult                    *megaport.DeleteMCRResponse
	CapturedDeleteMCRUID               string
	RestoreMCRErr                      error
	RestoreMCRResult                   *megaport.RestoreMCRResponse
	CapturedRestoreMCRUID              string
	ListMCRResourceTagsErr             error
	ListMCRResourceTagsResult          map[string]string
	CapturedListMCRResourceTagsUID     string
	UpdateMCRResourceTagsErr           error
	CapturedUpdateMCRResourceTagsUID   string
	CapturedUpdateMCRResourceTagsTags  map[string]string
	CreatePrefixFilterListErr          error
	CreatePrefixFilterListResult       *megaport.CreateMCRPrefixFilterListResponse
	CapturedCreatePrefixFilterListReq  *megaport.CreateMCRPrefixFilterListRequest
	ListMCRPrefixFilterListsErr        error
	ListMCRPrefixFilterListsResult     []*megaport.PrefixFilterList
	CapturedListMCRPrefixFilterListsID string
	GetMCRPrefixFilterListErr          error
	GetMCRPrefixFilterListResult       *megaport.MCRPrefixFilterList
	CapturedGetMCRPrefixFilterListID   string
	CapturedGetMCRPrefixFilterListNum  int
	ModifyMCRPrefixFilterListErr       error
	ModifyMCRPrefixFilterListResult    *megaport.ModifyMCRPrefixFilterListResponse
	CapturedModifyPrefixFilterListID   string
	CapturedModifyPrefixFilterListNum  int
	CapturedModifyPrefixFilterListReq  *megaport.MCRPrefixFilterList
	DeleteMCRPrefixFilterListErr       error
	DeleteMCRPrefixFilterListResult    *megaport.DeleteMCRPrefixFilterListResponse
	CapturedDeletePrefixFilterListID   string
	CapturedDeletePrefixFilterListNum  int
}

func (m *MockMCRService) BuyMCR(ctx context.Context, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	m.CapturedBuyMCRRequest = req
	if m.BuyMCRErr != nil {
		return nil, m.BuyMCRErr
	}
	if m.BuyMCRResult != nil {
		return m.BuyMCRResult, nil
	}
	return &megaport.BuyMCRResponse{
		TechnicalServiceUID: "mock-mcr-uid",
	}, nil
}

func (m *MockMCRService) ValidateMCROrder(ctx context.Context, req *megaport.BuyMCRRequest) error {
	if m.ValidateMCROrderErr != nil {
		return m.ValidateMCROrderErr
	}
	return nil
}

func (m *MockMCRService) GetMCR(ctx context.Context, mcrId string) (*megaport.MCR, error) {
	if m.GetMCRErr != nil {
		return nil, m.GetMCRErr
	}
	if m.GetMCRResult != nil {
		return m.GetMCRResult, nil
	}
	return &megaport.MCR{
		UID:                mcrId,
		Name:               "Mock MCR",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockMCRService) ListMCRs(ctx context.Context) ([]*megaport.MCR, error) {
	if m.ListMCRsErr != nil {
		return nil, m.ListMCRsErr
	}
	if m.ListMCRsResult != nil {
		return m.ListMCRsResult, nil
	}
	return []*megaport.MCR{}, nil
}

func (m *MockMCRService) ModifyMCR(ctx context.Context, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	m.CapturedModifyMCRRequest = req
	if m.ModifyMCRErr != nil {
		return nil, m.ModifyMCRErr
	}
	if m.ModifyMCRResult != nil {
		return m.ModifyMCRResult, nil
	}
	return &megaport.ModifyMCRResponse{
		IsUpdated: true,
	}, nil
}

func (m *MockMCRService) DeleteMCR(ctx context.Context, req *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	m.CapturedDeleteMCRUID = req.MCRID
	if m.DeleteMCRErr != nil {
		return nil, m.DeleteMCRErr
	}
	if m.DeleteMCRResult != nil {
		return m.DeleteMCRResult, nil
	}
	return &megaport.DeleteMCRResponse{
		IsDeleting: true,
	}, nil
}

func (m *MockMCRService) RestoreMCR(ctx context.Context, mcrId string) (*megaport.RestoreMCRResponse, error) {
	m.CapturedRestoreMCRUID = mcrId
	if m.RestoreMCRErr != nil {
		return nil, m.RestoreMCRErr
	}
	if m.RestoreMCRResult != nil {
		return m.RestoreMCRResult, nil
	}
	return &megaport.RestoreMCRResponse{
		IsRestored: true,
	}, nil
}

func (m *MockMCRService) ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	m.CapturedListMCRResourceTagsUID = mcrID
	if m.ListMCRResourceTagsErr != nil {
		return nil, m.ListMCRResourceTagsErr
	}
	if m.ListMCRResourceTagsResult != nil {
		return m.ListMCRResourceTagsResult, nil
	}
	return map[string]string{
		"environment": "test",
		"owner":       "automation",
	}, nil
}

func (m *MockMCRService) UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	m.CapturedUpdateMCRResourceTagsUID = mcrID
	m.CapturedUpdateMCRResourceTagsTags = tags
	if m.UpdateMCRResourceTagsErr != nil {
		return m.UpdateMCRResourceTagsErr
	}
	return nil
}

func (m *MockMCRService) CreatePrefixFilterList(ctx context.Context, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	m.CapturedCreatePrefixFilterListReq = req
	if m.CreatePrefixFilterListErr != nil {
		return nil, m.CreatePrefixFilterListErr
	}
	if m.CreatePrefixFilterListResult != nil {
		return m.CreatePrefixFilterListResult, nil
	}
	return &megaport.CreateMCRPrefixFilterListResponse{
		IsCreated:          true,
		PrefixFilterListID: 123,
	}, nil
}

func (m *MockMCRService) ListMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	m.CapturedListMCRPrefixFilterListsID = mcrId
	if m.ListMCRPrefixFilterListsErr != nil {
		return nil, m.ListMCRPrefixFilterListsErr
	}
	if m.ListMCRPrefixFilterListsResult != nil {
		return m.ListMCRPrefixFilterListsResult, nil
	}
	return []*megaport.PrefixFilterList{}, nil
}

func (m *MockMCRService) GetMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	m.CapturedGetMCRPrefixFilterListID = mcrID
	m.CapturedGetMCRPrefixFilterListNum = prefixFilterListID
	if m.GetMCRPrefixFilterListErr != nil {
		return nil, m.GetMCRPrefixFilterListErr
	}
	if m.GetMCRPrefixFilterListResult != nil {
		return m.GetMCRPrefixFilterListResult, nil
	}
	return &megaport.MCRPrefixFilterList{}, nil
}

func (m *MockMCRService) ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	m.CapturedModifyPrefixFilterListID = mcrID
	m.CapturedModifyPrefixFilterListNum = prefixFilterListID
	m.CapturedModifyPrefixFilterListReq = prefixFilterList
	if m.ModifyMCRPrefixFilterListErr != nil {
		return nil, m.ModifyMCRPrefixFilterListErr
	}
	if m.ModifyMCRPrefixFilterListResult != nil {
		return m.ModifyMCRPrefixFilterListResult, nil
	}
	return &megaport.ModifyMCRPrefixFilterListResponse{
		IsUpdated: true,
	}, nil
}

func (m *MockMCRService) DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	m.CapturedDeletePrefixFilterListID = mcrID
	m.CapturedDeletePrefixFilterListNum = prefixFilterListID
	if m.DeleteMCRPrefixFilterListErr != nil {
		return nil, m.DeleteMCRPrefixFilterListErr
	}
	if m.DeleteMCRPrefixFilterListResult != nil {
		return m.DeleteMCRPrefixFilterListResult, nil
	}
	return &megaport.DeleteMCRPrefixFilterListResponse{
		IsDeleted: true,
	}, nil
}

func (m *MockMCRService) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*megaport.PrefixFilterList, error) {
	return m.ListMCRPrefixFilterLists(ctx, mcrId)
}
