package users

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockUserManagementService struct {
	CreateUserErr    error
	CreateUserResult *megaport.CreateUserResponse
	GetUserErr       error
	GetUserResult    *megaport.User
	ListUsersErr     error
	ListUsersResult  []*megaport.User
	UpdateUserErr    error
	DeleteUserErr    error
	DeactivateErr    error
	ActivityErr      error
	ActivityResult   []*megaport.UserActivity

	CapturedCreateReq   *megaport.CreateUserRequest
	CapturedUpdateReq   *megaport.UpdateUserRequest
	CapturedEmployeeID  int
	CapturedActivityReq *megaport.GetUserActivityRequest
	ForceNilGetUser     bool
}

func (m *MockUserManagementService) CreateUser(ctx context.Context, req *megaport.CreateUserRequest) (*megaport.CreateUserResponse, error) {
	m.CapturedCreateReq = req
	if m.CreateUserErr != nil {
		return nil, m.CreateUserErr
	}
	if m.CreateUserResult != nil {
		return m.CreateUserResult, nil
	}
	return &megaport.CreateUserResponse{EmployeeID: 12345}, nil
}

func (m *MockUserManagementService) GetUser(ctx context.Context, employeeID int) (*megaport.User, error) {
	m.CapturedEmployeeID = employeeID
	if m.GetUserErr != nil {
		return nil, m.GetUserErr
	}
	if m.ForceNilGetUser {
		return nil, nil
	}
	if m.GetUserResult != nil {
		return m.GetUserResult, nil
	}
	return &megaport.User{
		PartyId:   employeeID,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Position:  "Technical Admin",
		Active:    true,
	}, nil
}

func (m *MockUserManagementService) ListCompanyUsers(ctx context.Context) ([]*megaport.User, error) {
	if m.ListUsersErr != nil {
		return nil, m.ListUsersErr
	}
	if m.ListUsersResult != nil {
		return m.ListUsersResult, nil
	}
	return []*megaport.User{}, nil
}

func (m *MockUserManagementService) UpdateUser(ctx context.Context, employeeID int, req *megaport.UpdateUserRequest) error {
	m.CapturedEmployeeID = employeeID
	m.CapturedUpdateReq = req
	return m.UpdateUserErr
}

func (m *MockUserManagementService) DeleteUser(ctx context.Context, employeeID int) error {
	m.CapturedEmployeeID = employeeID
	return m.DeleteUserErr
}

func (m *MockUserManagementService) DeactivateUser(ctx context.Context, employeeID int) error {
	m.CapturedEmployeeID = employeeID
	return m.DeactivateErr
}

func (m *MockUserManagementService) GetUserActivity(ctx context.Context, req *megaport.GetUserActivityRequest) ([]*megaport.UserActivity, error) {
	m.CapturedActivityReq = req
	if m.ActivityErr != nil {
		return nil, m.ActivityErr
	}
	if m.ActivityResult != nil {
		return m.ActivityResult, nil
	}
	return []*megaport.UserActivity{}, nil
}
