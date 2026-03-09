package billing_market

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockBillingMarketService struct {
	GetBillingMarketsError  error
	GetBillingMarketsResult []*megaport.BillingMarket

	SetBillingMarketError           error
	SetBillingMarketResult          *megaport.SetBillingMarketResponse
	CapturedSetBillingMarketRequest *megaport.SetBillingMarketRequest
}

func (m *MockBillingMarketService) GetBillingMarkets(ctx context.Context) ([]*megaport.BillingMarket, error) {
	if m.GetBillingMarketsError != nil {
		return nil, m.GetBillingMarketsError
	}
	if m.GetBillingMarketsResult != nil {
		return m.GetBillingMarketsResult, nil
	}
	return []*megaport.BillingMarket{
		{
			ID:                 1,
			SupplierName:       "Mock Supplier",
			CurrencyEnum:       "USD",
			Country:            "US",
			Region:             "North America",
			BillingContactName: "John Doe",
			Active:             true,
		},
	}, nil
}

func (m *MockBillingMarketService) SetBillingMarket(ctx context.Context, req *megaport.SetBillingMarketRequest) (*megaport.SetBillingMarketResponse, error) {
	m.CapturedSetBillingMarketRequest = req
	if m.SetBillingMarketError != nil {
		return nil, m.SetBillingMarketError
	}
	if m.SetBillingMarketResult != nil {
		return m.SetBillingMarketResult, nil
	}
	return &megaport.SetBillingMarketResponse{
		SupplyID: 12345,
	}, nil
}
