package managed_account

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type mockManagedAccountService struct {
	listResult             []*megaport.ManagedAccount
	listErr                error
	createResult           *megaport.ManagedAccount
	createErr              error
	capturedCreateReq      *megaport.ManagedAccountRequest
	updateResult           *megaport.ManagedAccount
	updateErr              error
	capturedUpdateUID      string
	capturedUpdateReq      *megaport.ManagedAccountRequest
	getResult              *megaport.ManagedAccount
	getErr                 error
	capturedGetCompanyUID  string
	capturedGetAccountName string
}

func (m *mockManagedAccountService) ListManagedAccounts(ctx context.Context) ([]*megaport.ManagedAccount, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listResult != nil {
		return m.listResult, nil
	}
	return []*megaport.ManagedAccount{}, nil
}

func (m *mockManagedAccountService) CreateManagedAccount(ctx context.Context, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
	m.capturedCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockManagedAccountService) UpdateManagedAccount(ctx context.Context, companyUID string, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
	m.capturedUpdateUID = companyUID
	m.capturedUpdateReq = req
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return m.updateResult, nil
}

func (m *mockManagedAccountService) GetManagedAccount(ctx context.Context, companyUID string, name string) (*megaport.ManagedAccount, error) {
	m.capturedGetCompanyUID = companyUID
	m.capturedGetAccountName = name
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResult, nil
}
