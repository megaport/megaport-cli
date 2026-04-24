package vxc

import (
	"context"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// // Mock VXC service for testing
type MockVXCService struct {
	BuyVXCResponse                       *megaport.BuyVXCResponse
	BuyVXCError                          error
	ValidateVXCOrderError                error
	GetVXCResponse                       *megaport.VXC
	GetVXCError                          error
	DeleteVXCError                       error
	UpdateVXCError                       error
	LookupPartnerPortsResponse           *megaport.LookupPartnerPortsResponse
	LookupPartnerPortsError              error
	ListPartnerPortsResponse             *megaport.ListPartnerPortsResponse
	ListPartnerPortsError                error
	CapturedUpdateVXCRequest             *megaport.UpdateVXCRequest
	UpdateVXCResponse                    *megaport.VXC
	ListVXCResponse                      []*megaport.VXC
	BuyVXCErr                            error
	ListVXCErr                           error
	CapturedListVXCsRequest              *megaport.ListVXCsRequest
	OnBuyVXC                             func(context.Context, *megaport.BuyVXCRequest)
	ListVXCResourceTagsErr               error
	ListVXCResourceTagsResult            map[string]string
	CapturedUpdateVXCResourceTagsRequest map[string]string
	CapturedListVXCResourceTagsUID       string
	UpdateVXCResourceTagsErr             error
	ForceNilGetVXC                       bool
}

func (m *MockVXCService) BuyVXC(ctx context.Context, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	// Call the OnBuyVXC callback if set
	if m.OnBuyVXC != nil {
		m.OnBuyVXC(ctx, req)
	}
	return m.BuyVXCResponse, m.BuyVXCError
}

func (m *MockVXCService) ValidateVXCOrder(ctx context.Context, req *megaport.BuyVXCRequest) error {
	return m.ValidateVXCOrderError
}

func (m *MockVXCService) GetVXC(ctx context.Context, id string) (*megaport.VXC, error) {
	if m.GetVXCError != nil {
		return nil, m.GetVXCError
	}
	if m.ForceNilGetVXC {
		return nil, nil
	}
	if m.GetVXCResponse != nil {
		return m.GetVXCResponse, nil
	}
	return &megaport.VXC{
		UID:                id,
		Name:               "Mock VXC",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockVXCService) DeleteVXC(ctx context.Context, id string, req *megaport.DeleteVXCRequest) error {
	return m.DeleteVXCError
}

func (m *MockVXCService) UpdateVXC(ctx context.Context, id string, req *megaport.UpdateVXCRequest) (*megaport.VXC, error) {
	m.CapturedUpdateVXCRequest = req
	return m.UpdateVXCResponse, m.UpdateVXCError
}

func (m *MockVXCService) ListVXCs(ctx context.Context, req *megaport.ListVXCsRequest) ([]*megaport.VXC, error) {
	m.CapturedListVXCsRequest = req
	if m.ListVXCErr != nil {
		return nil, m.ListVXCErr
	}
	// Simulate server-side filtering for the subset of request fields this mock
	// supports: NameContains, AEndProductUID, BEndProductUID, RateLimit, and
	// Status. Nils are passed through so callers handle them as they would with
	// a real SDK response.
	var result []*megaport.VXC
	for _, vxc := range m.ListVXCResponse {
		if vxc == nil {
			result = append(result, nil)
			continue
		}
		if req.NameContains != "" && !strings.Contains(vxc.Name, req.NameContains) {
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
	return m.LookupPartnerPortsResponse, m.LookupPartnerPortsError
}

func (m *MockVXCService) ListPartnerPorts(ctx context.Context, req *megaport.ListPartnerPortsRequest) (*megaport.ListPartnerPortsResponse, error) {
	return m.ListPartnerPortsResponse, m.ListPartnerPortsError
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

// Reset clears all configured results, errors, and captured requests.
func (m *MockVXCService) Reset() {
	m.BuyVXCResponse = nil
	m.BuyVXCError = nil
	m.ValidateVXCOrderError = nil
	m.GetVXCResponse = nil
	m.GetVXCError = nil
	m.DeleteVXCError = nil
	m.UpdateVXCError = nil
	m.LookupPartnerPortsResponse = nil
	m.LookupPartnerPortsError = nil
	m.ListPartnerPortsResponse = nil
	m.ListPartnerPortsError = nil
	m.CapturedUpdateVXCRequest = nil
	m.UpdateVXCResponse = nil
	m.ListVXCResponse = nil
	m.BuyVXCErr = nil
	m.ListVXCErr = nil
	m.CapturedListVXCsRequest = nil
	m.OnBuyVXC = nil
	m.ListVXCResourceTagsErr = nil
	m.ListVXCResourceTagsResult = nil
	m.CapturedUpdateVXCResourceTagsRequest = nil
	m.CapturedListVXCResourceTagsUID = ""
	m.UpdateVXCResourceTagsErr = nil
	m.ForceNilGetVXC = false
}
