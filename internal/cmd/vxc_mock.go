package cmd

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// // Mock VXC service for testing
type mockVXCService struct {
	buyVXCResponse              *megaport.BuyVXCResponse
	buyVXCError                 error
	validateVXCOrderError       error
	getVXCResponse              *megaport.VXC
	getVXCError                 error
	deleteVXCError              error
	updateVXCError              error
	lookupPartnerPortsResponse  *megaport.LookupPartnerPortsResponse
	lookupPartnerPortsError     error
	listPartnerPortsResponse    *megaport.ListPartnerPortsResponse
	listPartnerPortsError       error
	listVXCResourceTagsResponse map[string]string
	listVXCResourceTagsError    error
	updateVXCResourceTagsError  error
	updateVXCResponse           *megaport.VXC
	buyVXCErr                   error
	updateVXCErr                error
	onBuyVXC                    func(context.Context, *megaport.BuyVXCRequest)
	deleteVXCErr                error
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
	return m.listVXCResourceTagsResponse, m.listVXCResourceTagsError
}

func (m *mockVXCService) UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error {
	return m.updateVXCResourceTagsError
}
