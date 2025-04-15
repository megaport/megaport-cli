package ports

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// MockPortService implements the required Port service methods for testing
type MockPortService struct {
	// Optional fields to customize behavior
	GetPortErr                      error
	GetPortResult                   *megaport.Port
	ListPortsErr                    error
	ListPortsResult                 []*megaport.Port
	BuyPortErr                      error
	BuyPortResult                   *megaport.BuyPortResponse
	CapturedRequest                 *megaport.BuyPortRequest
	CheckPortVLANAvailabilityErr    error
	CheckPortVLANAvailabilityResult bool
	CapturedVLANRequest             struct {
		PortID string
		VLANID int
	}
	DeletePortErr         error
	DeletePortResult      *megaport.DeletePortResponse
	CapturedDeletePortUID string

	ListPortResourceTagsErr    error
	ListPortResourceTagsResult map[string]string
	CapturedResourceTagPortUID string
	CapturedResourceTags       map[string]string

	ValidatePortOrderErr      error
	ModifyPortErr             error
	ModifyPortResult          *megaport.ModifyPortResponse
	CapturedModifyPortRequest *megaport.ModifyPortRequest
	RestorePortErr            error
	RestorePortResult         *megaport.RestorePortResponse
	CapturedRestorePortUID    string
	LockPortErr               error
	LockPortResult            *megaport.LockPortResponse
	CapturedLockPortUID       string
	UnlockPortErr             error
	UnlockPortResult          *megaport.UnlockPortResponse
	CapturedUnlockPortUID     string
	UpdatePortResourceTagsErr error
	CapturedUpdateTagsRequest struct {
		PortID string
		Tags   map[string]string
	}
	UpdatePortErr    error
	UpdatePortResult *megaport.ModifyPortResponse
}

func (m *MockPortService) GetPort(ctx context.Context, portID string) (*megaport.Port, error) {
	if m.GetPortErr != nil {
		return nil, m.GetPortErr
	}
	if m.GetPortResult != nil {
		return m.GetPortResult, nil
	}
	return &megaport.Port{
		UID:                portID,
		Name:               "Mock Port",
		ProvisioningStatus: "LIVE",
	}, nil
}

func (m *MockPortService) ListPorts(ctx context.Context) ([]*megaport.Port, error) {
	if m.ListPortsErr != nil {
		return nil, m.ListPortsErr
	}
	if m.ListPortsResult != nil {
		return m.ListPortsResult, nil
	}
	return []*megaport.Port{}, nil
}

func (m *MockPortService) BuyPort(ctx context.Context, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	m.CapturedRequest = req

	if m.BuyPortErr != nil {
		return nil, m.BuyPortErr
	}
	if m.BuyPortResult != nil {
		return m.BuyPortResult, nil
	}
	return &megaport.BuyPortResponse{
		TechnicalServiceUIDs: []string{"mock-port-uid"},
	}, nil
}

func (m *MockPortService) CheckPortVLANAvailability(ctx context.Context, portID string, vlanID int) (bool, error) {
	m.CapturedVLANRequest.PortID = portID
	m.CapturedVLANRequest.VLANID = vlanID

	if m.CheckPortVLANAvailabilityErr != nil {
		return false, m.CheckPortVLANAvailabilityErr
	}

	return m.CheckPortVLANAvailabilityResult, nil
}

func (m *MockPortService) DeletePort(ctx context.Context, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	m.CapturedDeletePortUID = req.PortID

	if m.DeletePortErr != nil {
		return nil, m.DeletePortErr
	}

	if m.DeletePortResult != nil {
		return m.DeletePortResult, nil
	}

	return &megaport.DeletePortResponse{
		IsDeleting: true,
	}, nil
}

func (m *MockPortService) ListPortResourceTags(ctx context.Context, portID string) (map[string]string, error) {
	m.CapturedResourceTagPortUID = portID

	if m.ListPortResourceTagsErr != nil {
		return nil, m.ListPortResourceTagsErr
	}

	if m.ListPortResourceTagsResult != nil {
		return m.ListPortResourceTagsResult, nil
	}

	return map[string]string{
		"environment": "test",
		"owner":       "automation",
	}, nil
}

func (m *MockPortService) ValidatePortOrder(ctx context.Context, req *megaport.BuyPortRequest) error {
	if m.ValidatePortOrderErr != nil {
		return m.ValidatePortOrderErr
	}
	return nil
}

func (m *MockPortService) LockPort(ctx context.Context, portId string) (*megaport.LockPortResponse, error) {
	m.CapturedLockPortUID = portId

	if m.LockPortErr != nil {
		return nil, m.LockPortErr
	}

	if m.LockPortResult != nil {
		return m.LockPortResult, nil
	}

	return &megaport.LockPortResponse{
		IsLocking: true,
	}, nil
}

func (m *MockPortService) ModifyPort(ctx context.Context, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	m.CapturedModifyPortRequest = req

	if m.ModifyPortErr != nil {
		return nil, m.ModifyPortErr
	}

	if m.ModifyPortResult != nil {
		return m.ModifyPortResult, nil
	}

	return &megaport.ModifyPortResponse{
		IsUpdated: true,
	}, nil
}

func (m *MockPortService) RestorePort(ctx context.Context, portId string) (*megaport.RestorePortResponse, error) {
	m.CapturedRestorePortUID = portId

	if m.RestorePortErr != nil {
		return nil, m.RestorePortErr
	}

	if m.RestorePortResult != nil {
		return m.RestorePortResult, nil
	}

	return &megaport.RestorePortResponse{
		IsRestored: true,
	}, nil
}

func (m *MockPortService) UnlockPort(ctx context.Context, portId string) (*megaport.UnlockPortResponse, error) {
	m.CapturedUnlockPortUID = portId

	if m.UnlockPortErr != nil {
		return nil, m.UnlockPortErr
	}

	if m.UnlockPortResult != nil {
		return m.UnlockPortResult, nil
	}

	return &megaport.UnlockPortResponse{
		IsUnlocking: true,
	}, nil
}

// UpdatePortResourceTags implements the megaport.PortServicer interface
func (m *MockPortService) UpdatePortResourceTags(ctx context.Context, portID string, tags map[string]string) error {
	// Initialize the map if it's nil
	if m.CapturedResourceTags == nil {
		m.CapturedResourceTags = make(map[string]string)
	}

	// Copy the tags to the captured map
	for k, v := range tags {
		m.CapturedResourceTags[k] = v
	}

	return m.UpdatePortResourceTagsErr
}
