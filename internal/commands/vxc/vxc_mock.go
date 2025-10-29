package vxc

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// // Mock VXC service for testing
type mockVXCService struct {
	buyVXCResponse                       *megaport.BuyVXCResponse
	buyVXCError                          error
	validateVXCOrderError                error
	getVXCResponse                       *megaport.VXC
	getVXCError                          error
	deleteVXCError                       error
	updateVXCError                       error
	lookupPartnerPortsResponse           *megaport.LookupPartnerPortsResponse
	lookupPartnerPortsError              error
	listPartnerPortsResponse             *megaport.ListPartnerPortsResponse
	listPartnerPortsError                error
	updateVXCResponse                    *megaport.VXC
	buyVXCErr                            error
	onBuyVXC                             func(context.Context, *megaport.BuyVXCRequest)
	ListVXCResourceTagsErr               error
	ListVXCResourceTagsResult            map[string]string
	CapturedUpdateVXCResourceTagsRequest map[string]string
	CapturedListVXCResourceTagsUID       string
	UpdateVXCResourceTagsErr             error
}

func (m *mockVXCService) BuyVXC(ctx context.Context, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	// Call the onBuyVXC callback if set
	if m.onBuyVXC != nil {
		m.onBuyVXC(ctx, req)
	}
	return m.buyVXCResponse, m.buyVXCError
}

func (m *mockVXCService) ValidateVXCOrder(ctx context.Context, req *megaport.BuyVXCRequest) error {
	return m.validateVXCOrderError
}

func (m *mockVXCService) GetVXC(ctx context.Context, id string) (*megaport.VXC, error) {
	return m.getVXCResponse, m.getVXCError
}

func (m *mockVXCService) DeleteVXC(ctx context.Context, id string, req *megaport.DeleteVXCRequest) error {
	return m.deleteVXCError
}

func (m *mockVXCService) UpdateVXC(ctx context.Context, id string, req *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	return m.updateVXCResponse, m.updateVXCError
}

func (m *mockVXCService) LookupPartnerPorts(ctx context.Context, req *megaport.LookupPartnerPortsRequest) (*megaport.LookupPartnerPortsResponse, error) {
	return m.lookupPartnerPortsResponse, m.lookupPartnerPortsError
}

func (m *mockVXCService) ListPartnerPorts(ctx context.Context, req *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return m.listPartnerPortsResponse, m.listPartnerPortsError
}

func (m *mockVXCService) ListVXCResourceTags(ctx context.Context, vxcID string) (map[string]string, error) {
	if m.ListVXCResourceTagsErr != nil {
		return nil, m.ListVXCResourceTagsErr
	}
	return m.ListVXCResourceTagsResult, nil
}

func (m *mockVXCService) UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error {
	m.CapturedUpdateVXCResourceTagsRequest = tags
	return m.UpdateVXCResourceTagsErr
}

func (m *mockVXCService) ListVXCs(ctx context.Context, req *megaport.ListVXCsRequest) ([]*megaport.VXC, error) {
	// Return empty list if not implemented in test
	return []*megaport.VXC{}, nil
}
