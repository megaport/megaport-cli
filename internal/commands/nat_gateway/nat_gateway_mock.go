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

	ListPacketFiltersResult  []*megaport.NATGatewayPacketFilterSummary
	ListPacketFiltersErr     error
	CreatePacketFilterResult *megaport.NATGatewayPacketFilter
	CreatePacketFilterErr    error
	GetPacketFilterResult    *megaport.NATGatewayPacketFilter
	GetPacketFilterErr       error
	UpdatePacketFilterResult *megaport.NATGatewayPacketFilter
	UpdatePacketFilterErr    error
	DeletePacketFilterErr    error

	CapturedListPacketFiltersUID  string
	CapturedCreatePacketFilterUID string
	CapturedCreatePacketFilterReq *megaport.NATGatewayPacketFilterRequest
	CapturedGetPacketFilterUID    string
	CapturedGetPacketFilterID     int
	CapturedUpdatePacketFilterUID string
	CapturedUpdatePacketFilterID  int
	CapturedUpdatePacketFilterReq *megaport.NATGatewayPacketFilterRequest
	CapturedDeletePacketFilterUID string
	CapturedDeletePacketFilterID  int

	ListPrefixListsResult  []*megaport.NATGatewayPrefixListSummary
	ListPrefixListsErr     error
	CreatePrefixListResult *megaport.NATGatewayPrefixList
	CreatePrefixListErr    error
	GetPrefixListResult    *megaport.NATGatewayPrefixList
	GetPrefixListErr       error
	UpdatePrefixListResult *megaport.NATGatewayPrefixList
	UpdatePrefixListErr    error
	DeletePrefixListErr    error

	CapturedListPrefixListsUID  string
	CapturedCreatePrefixListUID string
	CapturedCreatePrefixListReq *megaport.NATGatewayPrefixList
	CapturedGetPrefixListUID    string
	CapturedGetPrefixListID     int
	CapturedUpdatePrefixListUID string
	CapturedUpdatePrefixListID  int
	CapturedUpdatePrefixListReq *megaport.NATGatewayPrefixList
	CapturedDeletePrefixListUID string
	CapturedDeletePrefixListID  int

	IPRoutesAsyncResult          string
	IPRoutesAsyncErr             error
	BGPRoutesAsyncResult         string
	BGPRoutesAsyncErr            error
	BGPNeighborRoutesAsyncResult string
	BGPNeighborRoutesAsyncErr    error
	DiagnosticsRoutesResult      []*megaport.NATGatewayRoute
	DiagnosticsRoutesErr         error
	IPRoutesResult               []*megaport.NATGatewayIPRoute
	IPRoutesErr                  error
	BGPRoutesResult              []*megaport.NATGatewayBGPRoute
	BGPRoutesErr                 error
	BGPNeighborRoutesResult      []*megaport.NATGatewayBGPRoute
	BGPNeighborRoutesErr         error

	CapturedIPRoutesAsyncUID          string
	CapturedIPRoutesAsyncIPAddr       string
	CapturedBGPRoutesAsyncUID         string
	CapturedBGPRoutesAsyncIPAddr      string
	CapturedBGPNeighborRoutesAsyncReq *megaport.NATGatewayBGPNeighborRoutesRequest
	CapturedDiagnosticsRoutesUID      string
	CapturedDiagnosticsRoutesOpID     string
	CapturedIPRoutesUID               string
	CapturedIPRoutesIPAddr            string
	CapturedBGPRoutesUID              string
	CapturedBGPRoutesIPAddr           string
	CapturedBGPNeighborRoutesReq      *megaport.NATGatewayBGPNeighborRoutesRequest
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
	m.CapturedListPacketFiltersUID = productUID
	if m.ListPacketFiltersErr != nil {
		return nil, m.ListPacketFiltersErr
	}
	return m.ListPacketFiltersResult, nil
}

func (m *MockNATGatewayService) CreateNATGatewayPacketFilter(ctx context.Context, productUID string, req *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	m.CapturedCreatePacketFilterUID = productUID
	m.CapturedCreatePacketFilterReq = req
	if m.CreatePacketFilterErr != nil {
		return nil, m.CreatePacketFilterErr
	}
	return m.CreatePacketFilterResult, nil
}

func (m *MockNATGatewayService) GetNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) (*megaport.NATGatewayPacketFilter, error) {
	m.CapturedGetPacketFilterUID = productUID
	m.CapturedGetPacketFilterID = packetFilterID
	if m.GetPacketFilterErr != nil {
		return nil, m.GetPacketFilterErr
	}
	return m.GetPacketFilterResult, nil
}

func (m *MockNATGatewayService) UpdateNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int, req *megaport.NATGatewayPacketFilterRequest) (*megaport.NATGatewayPacketFilter, error) {
	m.CapturedUpdatePacketFilterUID = productUID
	m.CapturedUpdatePacketFilterID = packetFilterID
	m.CapturedUpdatePacketFilterReq = req
	if m.UpdatePacketFilterErr != nil {
		return nil, m.UpdatePacketFilterErr
	}
	return m.UpdatePacketFilterResult, nil
}

func (m *MockNATGatewayService) DeleteNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) error {
	m.CapturedDeletePacketFilterUID = productUID
	m.CapturedDeletePacketFilterID = packetFilterID
	return m.DeletePacketFilterErr
}

func (m *MockNATGatewayService) ListNATGatewayPrefixLists(ctx context.Context, productUID string) ([]*megaport.NATGatewayPrefixListSummary, error) {
	m.CapturedListPrefixListsUID = productUID
	if m.ListPrefixListsErr != nil {
		return nil, m.ListPrefixListsErr
	}
	return m.ListPrefixListsResult, nil
}

func (m *MockNATGatewayService) CreateNATGatewayPrefixList(ctx context.Context, productUID string, req *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	m.CapturedCreatePrefixListUID = productUID
	m.CapturedCreatePrefixListReq = req
	if m.CreatePrefixListErr != nil {
		return nil, m.CreatePrefixListErr
	}
	return m.CreatePrefixListResult, nil
}

func (m *MockNATGatewayService) GetNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) (*megaport.NATGatewayPrefixList, error) {
	m.CapturedGetPrefixListUID = productUID
	m.CapturedGetPrefixListID = prefixListID
	if m.GetPrefixListErr != nil {
		return nil, m.GetPrefixListErr
	}
	return m.GetPrefixListResult, nil
}

func (m *MockNATGatewayService) UpdateNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int, req *megaport.NATGatewayPrefixList) (*megaport.NATGatewayPrefixList, error) {
	m.CapturedUpdatePrefixListUID = productUID
	m.CapturedUpdatePrefixListID = prefixListID
	m.CapturedUpdatePrefixListReq = req
	if m.UpdatePrefixListErr != nil {
		return nil, m.UpdatePrefixListErr
	}
	return m.UpdatePrefixListResult, nil
}

func (m *MockNATGatewayService) DeleteNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) error {
	m.CapturedDeletePrefixListUID = productUID
	m.CapturedDeletePrefixListID = prefixListID
	return m.DeletePrefixListErr
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	m.CapturedIPRoutesAsyncUID = productUID
	m.CapturedIPRoutesAsyncIPAddr = ipAddress
	return m.IPRoutesAsyncResult, m.IPRoutesAsyncErr
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	m.CapturedBGPRoutesAsyncUID = productUID
	m.CapturedBGPRoutesAsyncIPAddr = ipAddress
	return m.BGPRoutesAsyncResult, m.BGPRoutesAsyncErr
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutesAsync(ctx context.Context, req *megaport.NATGatewayBGPNeighborRoutesRequest) (string, error) {
	m.CapturedBGPNeighborRoutesAsyncReq = req
	return m.BGPNeighborRoutesAsyncResult, m.BGPNeighborRoutesAsyncErr
}

func (m *MockNATGatewayService) GetNATGatewayDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*megaport.NATGatewayRoute, error) {
	m.CapturedDiagnosticsRoutesUID = productUID
	m.CapturedDiagnosticsRoutesOpID = operationID
	if m.DiagnosticsRoutesErr != nil {
		return nil, m.DiagnosticsRoutesErr
	}
	return m.DiagnosticsRoutesResult, nil
}

func (m *MockNATGatewayService) ListNATGatewayIPRoutes(ctx context.Context, productUID, ipAddress string) ([]*megaport.NATGatewayIPRoute, error) {
	m.CapturedIPRoutesUID = productUID
	m.CapturedIPRoutesIPAddr = ipAddress
	if m.IPRoutesErr != nil {
		return nil, m.IPRoutesErr
	}
	return m.IPRoutesResult, nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPRoutes(ctx context.Context, productUID, ipAddress string) ([]*megaport.NATGatewayBGPRoute, error) {
	m.CapturedBGPRoutesUID = productUID
	m.CapturedBGPRoutesIPAddr = ipAddress
	if m.BGPRoutesErr != nil {
		return nil, m.BGPRoutesErr
	}
	return m.BGPRoutesResult, nil
}

func (m *MockNATGatewayService) ListNATGatewayBGPNeighborRoutes(ctx context.Context, req *megaport.NATGatewayBGPNeighborRoutesRequest) ([]*megaport.NATGatewayBGPRoute, error) {
	m.CapturedBGPNeighborRoutesReq = req
	if m.BGPNeighborRoutesErr != nil {
		return nil, m.BGPNeighborRoutesErr
	}
	return m.BGPNeighborRoutesResult, nil
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

	m.ListPacketFiltersResult = nil
	m.ListPacketFiltersErr = nil
	m.CreatePacketFilterResult = nil
	m.CreatePacketFilterErr = nil
	m.GetPacketFilterResult = nil
	m.GetPacketFilterErr = nil
	m.UpdatePacketFilterResult = nil
	m.UpdatePacketFilterErr = nil
	m.DeletePacketFilterErr = nil
	m.CapturedListPacketFiltersUID = ""
	m.CapturedCreatePacketFilterUID = ""
	m.CapturedCreatePacketFilterReq = nil
	m.CapturedGetPacketFilterUID = ""
	m.CapturedGetPacketFilterID = 0
	m.CapturedUpdatePacketFilterUID = ""
	m.CapturedUpdatePacketFilterID = 0
	m.CapturedUpdatePacketFilterReq = nil
	m.CapturedDeletePacketFilterUID = ""
	m.CapturedDeletePacketFilterID = 0

	m.ListPrefixListsResult = nil
	m.ListPrefixListsErr = nil
	m.CreatePrefixListResult = nil
	m.CreatePrefixListErr = nil
	m.GetPrefixListResult = nil
	m.GetPrefixListErr = nil
	m.UpdatePrefixListResult = nil
	m.UpdatePrefixListErr = nil
	m.DeletePrefixListErr = nil
	m.CapturedListPrefixListsUID = ""
	m.CapturedCreatePrefixListUID = ""
	m.CapturedCreatePrefixListReq = nil
	m.CapturedGetPrefixListUID = ""
	m.CapturedGetPrefixListID = 0
	m.CapturedUpdatePrefixListUID = ""
	m.CapturedUpdatePrefixListID = 0
	m.CapturedUpdatePrefixListReq = nil
	m.CapturedDeletePrefixListUID = ""
	m.CapturedDeletePrefixListID = 0

	m.IPRoutesAsyncResult = ""
	m.IPRoutesAsyncErr = nil
	m.BGPRoutesAsyncResult = ""
	m.BGPRoutesAsyncErr = nil
	m.BGPNeighborRoutesAsyncResult = ""
	m.BGPNeighborRoutesAsyncErr = nil
	m.DiagnosticsRoutesResult = nil
	m.DiagnosticsRoutesErr = nil
	m.IPRoutesResult = nil
	m.IPRoutesErr = nil
	m.BGPRoutesResult = nil
	m.BGPRoutesErr = nil
	m.BGPNeighborRoutesResult = nil
	m.BGPNeighborRoutesErr = nil
	m.CapturedIPRoutesAsyncUID = ""
	m.CapturedIPRoutesAsyncIPAddr = ""
	m.CapturedBGPRoutesAsyncUID = ""
	m.CapturedBGPRoutesAsyncIPAddr = ""
	m.CapturedBGPNeighborRoutesAsyncReq = nil
	m.CapturedDiagnosticsRoutesUID = ""
	m.CapturedDiagnosticsRoutesOpID = ""
	m.CapturedIPRoutesUID = ""
	m.CapturedIPRoutesIPAddr = ""
	m.CapturedBGPRoutesUID = ""
	m.CapturedBGPRoutesIPAddr = ""
	m.CapturedBGPNeighborRoutesReq = nil
}
