package nat_gateway

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

// MockNATGatewayService implements megaport.NATGatewayService for testing.
type MockNATGatewayService struct {
	CreateResult    *megaport.NATGateway
	CreateErr       error
	ListResult      []*megaport.NATGateway
	ListErr         error
	GetResult       *megaport.NATGateway
	GetErr          error
	UpdateResult    *megaport.NATGateway
	UpdateErr       error
	DeleteErr       error
	SessionsResult  []*megaport.NATGatewaySession
	SessionsErr     error
	TelemetryResult *megaport.ServiceTelemetryResponse
	TelemetryErr    error
	BuyResult       *megaport.NATGatewayBuyResult
	BuyErr          error
	ValidateResult  *megaport.NATGatewayValidateResult
	ValidateErr     error

	CapturedCreateReq    *megaport.CreateNATGatewayRequest
	CapturedUpdateReq    *megaport.UpdateNATGatewayRequest
	CapturedDeleteUID    string
	CapturedGetUID       string
	CapturedTelemetryReq *megaport.GetNATGatewayTelemetryRequest
	CapturedBuyUID       string
	CapturedValidateUID  string
}

func (m *MockNATGatewayService) CreateNATGateway(ctx context.Context, req *megaport.CreateNATGatewayRequest) (*megaport.NATGateway, error) {
	m.CapturedCreateReq = req
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	if m.CreateResult != nil {
		return m.CreateResult, nil
	}
	return &megaport.NATGateway{ProductUID: "nat-uid-mock", ProductName: req.ProductName}, nil
}

func (m *MockNATGatewayService) ListNATGateways(ctx context.Context) ([]*megaport.NATGateway, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.ListResult, nil
}

func (m *MockNATGatewayService) GetNATGateway(ctx context.Context, productUID string) (*megaport.NATGateway, error) {
	m.CapturedGetUID = productUID
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	if m.GetResult != nil {
		return m.GetResult, nil
	}
	return &megaport.NATGateway{ProductUID: productUID}, nil
}

func (m *MockNATGatewayService) UpdateNATGateway(ctx context.Context, req *megaport.UpdateNATGatewayRequest) (*megaport.NATGateway, error) {
	m.CapturedUpdateReq = req
	if m.UpdateErr != nil {
		return nil, m.UpdateErr
	}
	if m.UpdateResult != nil {
		return m.UpdateResult, nil
	}
	return &megaport.NATGateway{ProductUID: req.ProductUID, ProductName: req.ProductName}, nil
}

func (m *MockNATGatewayService) DeleteNATGateway(ctx context.Context, productUID string) error {
	m.CapturedDeleteUID = productUID
	return m.DeleteErr
}

func (m *MockNATGatewayService) ListNATGatewaySessions(ctx context.Context) ([]*megaport.NATGatewaySession, error) {
	if m.SessionsErr != nil {
		return nil, m.SessionsErr
	}
	return m.SessionsResult, nil
}

func (m *MockNATGatewayService) GetNATGatewayTelemetry(ctx context.Context, req *megaport.GetNATGatewayTelemetryRequest) (*megaport.ServiceTelemetryResponse, error) {
	m.CapturedTelemetryReq = req
	if m.TelemetryErr != nil {
		return nil, m.TelemetryErr
	}
	if m.TelemetryResult != nil {
		return m.TelemetryResult, nil
	}
	return &megaport.ServiceTelemetryResponse{}, nil
}

func (m *MockNATGatewayService) ValidateNATGatewayOrder(ctx context.Context, productUID string) (*megaport.NATGatewayValidateResult, error) {
	m.CapturedValidateUID = productUID
	if m.ValidateErr != nil {
		return nil, m.ValidateErr
	}
	if m.ValidateResult != nil {
		return m.ValidateResult, nil
	}
	return &megaport.NATGatewayValidateResult{ProductUID: productUID}, nil
}

func (m *MockNATGatewayService) BuyNATGateway(ctx context.Context, productUID string) (*megaport.NATGatewayBuyResult, error) {
	m.CapturedBuyUID = productUID
	if m.BuyErr != nil {
		return nil, m.BuyErr
	}
	if m.BuyResult != nil {
		return m.BuyResult, nil
	}
	return &megaport.NATGatewayBuyResult{ProductUID: productUID}, nil
}

func (m *MockNATGatewayService) CreateNATGatewayPacketFilter(_ context.Context, _ string, _ *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	return nil, fmt.Errorf("mock: CreateNATGatewayPacketFilter not configured")
}

func (m *MockNATGatewayService) ListNATGatewayPacketFilters(_ context.Context, _ string) ([]*megaport.NATGatewayPacketFilterSummary, error) {
	return nil, fmt.Errorf("mock: ListNATGatewayPacketFilters not configured")
}

func (m *MockNATGatewayService) GetNATGatewayPacketFilter(_ context.Context, _ string, _ int) (*megaport.NATGatewayPacketFilter, error) {
	return nil, fmt.Errorf("mock: GetNATGatewayPacketFilter not configured")
}

func (m *MockNATGatewayService) UpdateNATGatewayPacketFilter(_ context.Context, _ string, _ int, _ *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	return nil, fmt.Errorf("mock: UpdateNATGatewayPacketFilter not configured")
}

func (m *MockNATGatewayService) DeleteNATGatewayPacketFilter(_ context.Context, _ string, _ int) error {
	return fmt.Errorf("mock: DeleteNATGatewayPacketFilter not configured")
}

func (m *MockNATGatewayService) ListNATGatewayPrefixLists(_ context.Context, _ string) ([]*megaport.NATGatewayPrefixListSummary, error) {
	return nil, fmt.Errorf("mock: ListNATGatewayPrefixLists not configured")
}

func (m *MockNATGatewayService) CreateNATGatewayPrefixList(_ context.Context, _ string, _ *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	return nil, fmt.Errorf("mock: CreateNATGatewayPrefixList not configured")
}

func (m *MockNATGatewayService) GetNATGatewayPrefixList(_ context.Context, _ string, _ int) (*megaport.NATGatewayPrefixList, error) {
	return nil, fmt.Errorf("mock: GetNATGatewayPrefixList not configured")
}

func (m *MockNATGatewayService) UpdateNATGatewayPrefixList(_ context.Context, _ string, _ int, _ *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	return nil, fmt.Errorf("mock: UpdateNATGatewayPrefixList not configured")
}

func (m *MockNATGatewayService) DeleteNATGatewayPrefixList(_ context.Context, _ string, _ int) error {
	return fmt.Errorf("mock: DeleteNATGatewayPrefixList not configured")
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutesAsync(_ context.Context, _, _ string) (string, error) {
	return "", fmt.Errorf("mock: ListNATGatewayIPRoutesAsync not configured")
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutesAsync(_ context.Context, _, _ string) (string, error) {
	return "", fmt.Errorf("mock: ListNATGatewayBGPRoutesAsync not configured")
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutesAsync(_ context.Context, _ *megaport.NATGatewayBGPNeighborRoutesRequest) (string, error) {
	return "", fmt.Errorf("mock: ListNATGatewayBGPNeighborRoutesAsync not configured")
}

func (m *MockNATGatewayService) GetNATGatewayDiagnosticsRoutes(_ context.Context, _, _ string) ([]*megaport.NATGatewayRoute, error) {
	return nil, fmt.Errorf("mock: GetNATGatewayDiagnosticsRoutes not configured")
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutes(_ context.Context, _, _ string) ([]*megaport.NATGatewayIPRoute, error) {
	return nil, fmt.Errorf("mock: ListNATGatewayIPRoutes not configured")
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutes(_ context.Context, _, _ string) ([]*megaport.NATGatewayBGPRoute, error) {
	return nil, fmt.Errorf("mock: ListNATGatewayBGPRoutes not configured")
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutes(_ context.Context, _ *megaport.NATGatewayBGPNeighborRoutesRequest) ([]*megaport.NATGatewayBGPRoute, error) {
	return nil, fmt.Errorf("mock: ListNATGatewayBGPNeighborRoutes not configured")
}

func (m *MockNATGatewayService) Reset() {
	m.CreateResult = nil
	m.CreateErr = nil
	m.ListResult = nil
	m.ListErr = nil
	m.GetResult = nil
	m.GetErr = nil
	m.UpdateResult = nil
	m.UpdateErr = nil
	m.DeleteErr = nil
	m.SessionsResult = nil
	m.SessionsErr = nil
	m.TelemetryResult = nil
	m.TelemetryErr = nil
	m.BuyResult = nil
	m.BuyErr = nil
	m.ValidateResult = nil
	m.ValidateErr = nil
	m.CapturedBuyUID = ""
	m.CapturedValidateUID = ""
	m.CapturedCreateReq = nil
	m.CapturedUpdateReq = nil
	m.CapturedDeleteUID = ""
	m.CapturedGetUID = ""
	m.CapturedTelemetryReq = nil
}
