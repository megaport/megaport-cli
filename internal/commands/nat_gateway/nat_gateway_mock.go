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

	CapturedCreateReq    *megaport.CreateNATGatewayRequest
	CapturedUpdateReq    *megaport.UpdateNATGatewayRequest
	CapturedDeleteUID    string
	CapturedGetUID       string
	CapturedTelemetryReq *megaport.GetNATGatewayTelemetryRequest
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
	m.CapturedCreateReq = nil
	m.CapturedUpdateReq = nil
	m.CapturedDeleteUID = ""
	m.CapturedGetUID = ""
	m.CapturedTelemetryReq = nil
}
