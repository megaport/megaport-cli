package servicekeys

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockServiceKeyService struct {
	CreateServiceKeyError           error
	CreateServiceKeyResult          *megaport.CreateServiceKeyResponse
	CapturedCreateServiceKeyRequest *megaport.CreateServiceKeyRequest

	ListServiceKeysError           error
	ListServiceKeysResult          *megaport.ListServiceKeysResponse
	CapturedListServiceKeysRequest *megaport.ListServiceKeysRequest

	UpdateServiceKeyError           error
	UpdateServiceKeyResult          *megaport.UpdateServiceKeyResponse
	CapturedUpdateServiceKeyRequest *megaport.UpdateServiceKeyRequest

	GetServiceKeyError         error
	GetServiceKeyResult        *megaport.ServiceKey
	CapturedGetServiceKeyKeyID string
}

func (m *MockServiceKeyService) CreateServiceKey(ctx context.Context, req *megaport.CreateServiceKeyRequest) (*megaport.CreateServiceKeyResponse, error) {
	m.CapturedCreateServiceKeyRequest = req
	if m.CreateServiceKeyError != nil {
		return nil, m.CreateServiceKeyError
	}
	if m.CreateServiceKeyResult != nil {
		return m.CreateServiceKeyResult, nil
	}
	return &megaport.CreateServiceKeyResponse{
		ServiceKeyUID: "mock-service-key-uid",
	}, nil
}

func (m *MockServiceKeyService) ListServiceKeys(ctx context.Context, req *megaport.ListServiceKeysRequest) (*megaport.ListServiceKeysResponse, error) {
	m.CapturedListServiceKeysRequest = req
	if m.ListServiceKeysError != nil {
		return nil, m.ListServiceKeysError
	}
	if m.ListServiceKeysResult != nil {
		return m.ListServiceKeysResult, nil
	}
	return &megaport.ListServiceKeysResponse{
		ServiceKeys: []*megaport.ServiceKey{
			{
				Key:         "mock-service-key-1",
				Description: "Mock Service Key 1",
				ProductUID:  "mock-product-uid-1",
				MaxSpeed:    1000,
				SingleUse:   true,
				Active:      true,
			},
			{
				Key:         "mock-service-key-2",
				Description: "Mock Service Key 2",
				ProductUID:  "mock-product-uid-2",
				MaxSpeed:    2000,
				SingleUse:   false,
				Active:      true,
			},
		},
	}, nil
}

func (m *MockServiceKeyService) UpdateServiceKey(ctx context.Context, req *megaport.UpdateServiceKeyRequest) (*megaport.UpdateServiceKeyResponse, error) {
	m.CapturedUpdateServiceKeyRequest = req
	if m.UpdateServiceKeyError != nil {
		return nil, m.UpdateServiceKeyError
	}
	if m.UpdateServiceKeyResult != nil {
		return m.UpdateServiceKeyResult, nil
	}
	return &megaport.UpdateServiceKeyResponse{
		IsUpdated: true,
	}, nil
}

func (m *MockServiceKeyService) GetServiceKey(ctx context.Context, keyId string) (*megaport.ServiceKey, error) {
	m.CapturedGetServiceKeyKeyID = keyId
	if m.GetServiceKeyError != nil {
		return nil, m.GetServiceKeyError
	}
	if m.GetServiceKeyResult != nil {
		return m.GetServiceKeyResult, nil
	}
	return &megaport.ServiceKey{
		Key:         keyId,
		Description: "Mock Service Key",
		ProductUID:  "mock-product-uid",
		MaxSpeed:    1000,
		SingleUse:   true,
		Active:      true,
	}, nil
}
