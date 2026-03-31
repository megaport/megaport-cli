package ix

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockIXService struct {
	buyIXResponse          *megaport.BuyIXResponse
	buyIXError             error
	capturedBuyIXRequest   *megaport.BuyIXRequest
	validateIXOrderError   error
	getIXResponse          *megaport.IX
	getIXError             error
	deleteIXError          error
	capturedDeleteIXUID    string
	capturedDeleteIXReq    *megaport.DeleteIXRequest
	updateIXResponse       *megaport.IX
	updateIXError          error
	capturedUpdateIXUID    string
	capturedUpdateIXReq    *megaport.UpdateIXRequest
	listIXResponse         []*megaport.IX
	listIXErr              error
	CapturedListIXsRequest *megaport.ListIXsRequest
	forceNilGetIX          bool
}

func (m *MockIXService) BuyIX(ctx context.Context, req *megaport.BuyIXRequest) (*megaport.BuyIXResponse, error) {
	m.capturedBuyIXRequest = req
	if m.buyIXError != nil {
		return nil, m.buyIXError
	}
	return m.buyIXResponse, nil
}

func (m *MockIXService) ValidateIXOrder(ctx context.Context, req *megaport.BuyIXRequest) error {
	return m.validateIXOrderError
}

func (m *MockIXService) GetIX(ctx context.Context, id string) (*megaport.IX, error) {
	if m.getIXError != nil {
		return nil, m.getIXError
	}
	if m.forceNilGetIX {
		return nil, nil
	}
	if m.getIXResponse != nil {
		return m.getIXResponse, nil
	}
	return &megaport.IX{
		ProductUID:         id,
		ProductName:        "Mock IX",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockIXService) UpdateIX(ctx context.Context, id string, req *megaport.UpdateIXRequest) (*megaport.IX, error) {
	m.capturedUpdateIXUID = id
	m.capturedUpdateIXReq = req
	if m.updateIXError != nil {
		return nil, m.updateIXError
	}
	return m.updateIXResponse, nil
}

func (m *MockIXService) DeleteIX(ctx context.Context, id string, req *megaport.DeleteIXRequest) error {
	m.capturedDeleteIXUID = id
	m.capturedDeleteIXReq = req
	return m.deleteIXError
}

func (m *MockIXService) ListIXs(ctx context.Context, req *megaport.ListIXsRequest) ([]*megaport.IX, error) {
	m.CapturedListIXsRequest = req
	if m.listIXErr != nil {
		return nil, m.listIXErr
	}
	if m.listIXResponse != nil {
		return m.listIXResponse, nil
	}
	return []*megaport.IX{}, nil
}
