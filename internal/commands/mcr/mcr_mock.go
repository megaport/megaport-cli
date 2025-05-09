package mcr

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockMCRService struct {
	BuyMCRResult                             *megaport.BuyMCRResponse
	BuyMCRErr                                error
	CapturedBuyMCRRequest                    *megaport.BuyMCRRequest
	ValidateMCROrderErr                      error
	GetMCRResult                             *megaport.MCR
	GetMCRErr                                error
	ListMCRsResult                           []*megaport.MCR
	ListMCRsErr                              error
	CapturedListMCRsRequest                  *megaport.ListMCRsRequest
	DeleteMCRResult                          *megaport.DeleteMCRResponse
	DeleteMCRErr                             error
	CapturedDeleteMCRUID                     string
	CapturedMCRUID                           string
	RestoreMCRResult                         *megaport.RestoreMCRResponse
	RestoreMCRErr                            error
	CapturedRestoreMCRUID                    string
	CreateMCRPrefixFilterListResult          *megaport.CreateMCRPrefixFilterListResponse
	CreateMCRPrefixFilterListErr             error
	CapturedCreateMCRPrefixFilterListRequest *megaport.CreateMCRPrefixFilterListRequest
	CapturedCreatePrefixFilterListRequest    *megaport.CreateMCRPrefixFilterListRequest

	CreateMCRPrefixFilterListResponse *megaport.CreateMCRPrefixFilterListResponse
	CreateMCRPrefixFilterListError    error
	CapturedCreatePrefixFilterList    *megaport.CreateMCRPrefixFilterListRequest

	ListMCRPrefixFilterListsResult           []*megaport.PrefixFilterList
	ListMCRPrefixFilterListsErr              error
	GetMCRPrefixFilterListResult             *megaport.MCRPrefixFilterList
	GetMCRPrefixFilterListErr                error
	ModifyMCRPrefixFilterListResult          *megaport.ModifyMCRPrefixFilterListResponse
	ModifyMCRPrefixFilterListErr             error
	CapturedModifyMCRPrefixFilterListRequest *megaport.MCRPrefixFilterList
	DeleteMCRPrefixFilterListResult          *megaport.DeleteMCRPrefixFilterListResponse
	DeleteMCRPrefixFilterListErr             error
	ModifyMCRResult                          *megaport.ModifyMCRResponse
	ModifyMCRErr                             error
	CapturedModifyMCRRequest                 *megaport.ModifyMCRRequest
	ListMCRResourceTagsResult                map[string]string
	ListMCRResourceTagsErr                   error
	UpdateMCRResourceTagsErr                 error
	CapturedUpdateMCRResourceTagsRequest     map[string]string
	GetMCRPrefixFilterListsResult            []*megaport.PrefixFilterList
	GetMCRPrefixFilterListsErr               error
	CapturedModifyPrefixFilterListMCRID      string
	CapturedModifyPrefixFilterListID         int
	CapturedModifyPrefixFilterList           *megaport.MCRPrefixFilterList
}

func (m *MockMCRService) BuyMCR(ctx context.Context, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	m.CapturedBuyMCRRequest = req
	if m.BuyMCRErr != nil {
		return nil, m.BuyMCRErr
	}
	return m.BuyMCRResult, nil
}

func (m *MockMCRService) ValidateMCROrder(ctx context.Context, req *megaport.BuyMCRRequest) error {
	return m.ValidateMCROrderErr
}

func (m *MockMCRService) GetMCR(ctx context.Context, mcrUID string) (*megaport.MCR, error) {
	if m.GetMCRErr != nil {
		return nil, m.GetMCRErr
	}
	return m.GetMCRResult, nil
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

func (m *MockMCRService) DeleteMCR(ctx context.Context, req *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	m.CapturedDeleteMCRUID = req.MCRID
	if m.DeleteMCRErr != nil {
		return nil, m.DeleteMCRErr
	}
	return m.DeleteMCRResult, nil
}

func (m *MockMCRService) RestoreMCR(ctx context.Context, mcrUID string) (*megaport.RestoreMCRResponse, error) {
	m.CapturedRestoreMCRUID = mcrUID
	if m.RestoreMCRErr != nil {
		return nil, m.RestoreMCRErr
	}
	return m.RestoreMCRResult, nil
}

func (m *MockMCRService) CreatePrefixFilterList(ctx context.Context, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	m.CapturedCreatePrefixFilterListRequest = req
	if m.CreateMCRPrefixFilterListErr != nil {
		return nil, m.CreateMCRPrefixFilterListErr
	}
	return m.CreateMCRPrefixFilterListResult, nil
}

func (m *MockMCRService) ListMCRPrefixFilterLists(ctx context.Context, mcrID string) ([]*megaport.PrefixFilterList, error) {
	if m.ListMCRPrefixFilterListsErr != nil {
		return nil, m.ListMCRPrefixFilterListsErr
	}
	return m.ListMCRPrefixFilterListsResult, nil
}

func (m *MockMCRService) GetMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	if m.GetMCRPrefixFilterListErr != nil {
		return nil, m.GetMCRPrefixFilterListErr
	}
	return m.GetMCRPrefixFilterListResult, nil
}

func (m *MockMCRService) ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	m.CapturedModifyPrefixFilterListMCRID = mcrID
	m.CapturedModifyPrefixFilterListID = prefixFilterListID
	m.CapturedModifyPrefixFilterList = prefixFilterList
	if m.ModifyMCRPrefixFilterListErr != nil {
		return nil, m.ModifyMCRPrefixFilterListErr
	}
	return m.ModifyMCRPrefixFilterListResult, nil
}

func (m *MockMCRService) DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	if m.DeleteMCRPrefixFilterListErr != nil {
		return nil, m.DeleteMCRPrefixFilterListErr
	}
	return m.DeleteMCRPrefixFilterListResult, nil
}

func (m *MockMCRService) ModifyMCR(ctx context.Context, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	m.CapturedModifyMCRRequest = req
	if m.ModifyMCRErr != nil {
		return nil, m.ModifyMCRErr
	}
	return m.ModifyMCRResult, nil
}

func (m *MockMCRService) ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	if m.ListMCRResourceTagsErr != nil {
		return nil, m.ListMCRResourceTagsErr
	}
	return m.ListMCRResourceTagsResult, nil
}

func (m *MockMCRService) UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	m.CapturedUpdateMCRResourceTagsRequest = tags
	return m.UpdateMCRResourceTagsErr
}

func (m *MockMCRService) GetMCRPrefixFilterLists(ctx context.Context, mcrID string) ([]*megaport.PrefixFilterList, error) {
	if m.GetMCRPrefixFilterListsErr != nil {
		return nil, m.GetMCRPrefixFilterListsErr
	}
	return m.GetMCRPrefixFilterListsResult, nil
}

func (m *MockMCRService) Reset() {
	m.BuyMCRResult = nil
	m.BuyMCRErr = nil
	m.CapturedBuyMCRRequest = nil
	m.ValidateMCROrderErr = nil
	m.GetMCRResult = nil
	m.GetMCRErr = nil
	m.ListMCRsResult = nil
	m.ListMCRsErr = nil
	m.CapturedListMCRsRequest = nil
	m.DeleteMCRResult = nil
	m.DeleteMCRErr = nil
	m.CapturedDeleteMCRUID = ""
	m.CapturedMCRUID = ""
	m.RestoreMCRResult = nil
	m.RestoreMCRErr = nil
	m.CapturedRestoreMCRUID = ""
	m.CreateMCRPrefixFilterListResult = nil
	m.CreateMCRPrefixFilterListErr = nil
	m.CapturedCreateMCRPrefixFilterListRequest = nil
	m.CapturedCreatePrefixFilterListRequest = nil
	m.CreateMCRPrefixFilterListResponse = nil
	m.CreateMCRPrefixFilterListError = nil
	m.CapturedCreatePrefixFilterList = nil
	m.ListMCRPrefixFilterListsResult = nil
	m.ListMCRPrefixFilterListsErr = nil
	m.GetMCRPrefixFilterListResult = nil
	m.GetMCRPrefixFilterListErr = nil
	m.ModifyMCRPrefixFilterListResult = nil
	m.ModifyMCRPrefixFilterListErr = nil
	m.CapturedModifyMCRPrefixFilterListRequest = nil
	m.DeleteMCRPrefixFilterListResult = nil
	m.DeleteMCRPrefixFilterListErr = nil
	m.ModifyMCRResult = nil
	m.ModifyMCRErr = nil
	m.CapturedModifyMCRRequest = nil
	m.ListMCRResourceTagsResult = nil
	m.ListMCRResourceTagsErr = nil
	m.UpdateMCRResourceTagsErr = nil
	m.CapturedUpdateMCRResourceTagsRequest = nil
	m.GetMCRPrefixFilterListsResult = nil
	m.GetMCRPrefixFilterListsErr = nil
	m.CapturedModifyPrefixFilterListMCRID = ""
	m.CapturedModifyPrefixFilterListID = 0
	m.CapturedModifyPrefixFilterList = nil
}
