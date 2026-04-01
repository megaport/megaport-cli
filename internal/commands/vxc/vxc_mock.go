package vxc

import (
	"context"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// // Mock VXC service for testing
type MockVXCService struct {
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
	capturedUpdateVXCRequest             *megaport.UpdateVXCRequest
	updateVXCResponse                    *megaport.VXC
	listVXCResponse                      []*megaport.VXC
	buyVXCErr                            error
	listVXCErr                           error
	CapturedListVXCsRequest              *megaport.ListVXCsRequest
	onBuyVXC                             func(context.Context, *megaport.BuyVXCRequest)
	ListVXCResourceTagsErr               error
	ListVXCResourceTagsResult            map[string]string
	CapturedUpdateVXCResourceTagsRequest map[string]string
	CapturedListVXCResourceTagsUID       string
	UpdateVXCResourceTagsErr             error
	forceNilGetVXC                       bool
}

func (m *MockVXCService) BuyVXC(ctx context.Context, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	// Call the onBuyVXC callback if set
	if m.onBuyVXC != nil {
		m.onBuyVXC(ctx, req)
	}
	return m.buyVXCResponse, m.buyVXCError
}

func (m *MockVXCService) ValidateVXCOrder(ctx context.Context, req *megaport.BuyVXCRequest) error {
	return m.validateVXCOrderError
}

func (m *MockVXCService) GetVXC(ctx context.Context, id string) (*megaport.VXC, error) {
	if m.getVXCError != nil {
		return nil, m.getVXCError
	}
	if m.forceNilGetVXC {
		return nil, nil
	}
	if m.getVXCResponse != nil {
		return m.getVXCResponse, nil
	}
	return &megaport.VXC{
		UID:                id,
		Name:               "Mock VXC",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockVXCService) DeleteVXC(ctx context.Context, id string, req *megaport.DeleteVXCRequest) error {
	return m.deleteVXCError
}

func (m *MockVXCService) UpdateVXC(ctx context.Context, id string, req *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	m.capturedUpdateVXCRequest = req
	return m.updateVXCResponse, m.updateVXCError
}

func (m *MockVXCService) ListVXCs(ctx context.Context, req *megaport.ListVXCsRequest) ([]*megaport.VXC, error) {
	m.CapturedListVXCsRequest = req
	if m.listVXCErr != nil {
		return nil, m.listVXCErr
	}
	// Simulate server-side filtering based on request fields
	var result []*megaport.VXC
	for _, vxc := range m.listVXCResponse {
		if vxc == nil {
			continue
		}
		if req.NameContains != "" && !strings.Contains(strings.ToLower(vxc.Name), strings.ToLower(req.NameContains)) {
			continue
		}
		if req.AEndProductUID != "" && vxc.AEndConfiguration.UID != req.AEndProductUID {
			continue
		}
		if req.BEndProductUID != "" && vxc.BEndConfiguration.UID != req.BEndProductUID {
			continue
		}
		if req.RateLimit > 0 && vxc.RateLimit != req.RateLimit {
			continue
		}
		if len(req.Status) > 0 {
			matched := false
			for _, s := range req.Status {
				if vxc.ProvisioningStatus == s {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		result = append(result, vxc)
	}
	return result, nil
}

func (m *MockVXCService) LookupPartnerPorts(ctx context.Context, req *megaport.LookupPartnerPortsRequest) (*megaport.LookupPartnerPortsResponse, error) {
	return m.lookupPartnerPortsResponse, m.lookupPartnerPortsError
}

func (m *MockVXCService) ListPartnerPorts(ctx context.Context, req *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return m.listPartnerPortsResponse, m.listPartnerPortsError
}

func (m *MockVXCService) ListVXCResourceTags(ctx context.Context, vxcID string) (map[string]string, error) {
	if m.ListVXCResourceTagsErr != nil {
		return nil, m.ListVXCResourceTagsErr
	}
	return m.ListVXCResourceTagsResult, nil
}

func (m *MockVXCService) UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error {
	m.CapturedUpdateVXCResourceTagsRequest = tags
	return m.UpdateVXCResourceTagsErr
}
