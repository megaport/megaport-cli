package nat_gateway

import (
	"context"

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

func (m *MockNATGatewayService) ListNATGatewayPacketFilters(ctx context.Context, productUID string) ([]*megaport.NATGatewayPacketFilterSummary, error) {
	return nil, nil
}

func (m *MockNATGatewayService) CreateNATGatewayPacketFilter(ctx context.Context, productUID string, req *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	return nil, nil
}

func (m *MockNATGatewayService) GetNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) (*megaport.NATGatewayPacketFilter, error) {
	return nil, nil
}

func (m *MockNATGatewayService) UpdateNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int, req *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	return nil, nil
}

func (m *MockNATGatewayService) DeleteNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) error {
	return nil
}

func (m *MockNATGatewayService) ListNATGatewayPrefixLists(ctx context.Context, productUID string) ([]*megaport.NATGatewayPrefixListSummary, error) {
	return nil, nil
}

func (m *MockNATGatewayService) CreateNATGatewayPrefixList(ctx context.Context, productUID string, req *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	return nil, nil
}

func (m *MockNATGatewayService) GetNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) (*megaport.NATGatewayPrefixList, error) {
	return nil, nil
}

func (m *MockNATGatewayService) UpdateNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int, req *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	return nil, nil
}

func (m *MockNATGatewayService) DeleteNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) error {
	return nil
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	return "", nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	return "", nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutesAsync(ctx context.Context, req *megaport.NATGatewayBGPNeighborRoutesRequest) (string, error) {
	return "", nil
}

func (m *MockNATGatewayService) GetNATGatewayDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*megaport.NATGatewayRoute, error) {
	return nil, nil
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutes(ctx context.Context, productUID, ipAddress string) ([]*megaport.NATGatewayIPRoute, error) {
	return nil, nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutes(ctx context.Context, productUID, ipAddress string) ([]*megaport.NATGatewayBGPRoute, error) {
	return nil, nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutes(ctx context.Context, req *megaport.NATGatewayBGPNeighborRoutesRequest) ([]*megaport.NATGatewayBGPRoute, error) {
	return nil, nil
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
	m.CapturedCreateReq = nil
	m.CapturedUpdateReq = nil
	m.CapturedDeleteUID = ""
	m.CapturedGetUID = ""
	m.CapturedTelemetryReq = nil
	m.CapturedBuyUID = ""
	m.CapturedValidateUID = ""
}
