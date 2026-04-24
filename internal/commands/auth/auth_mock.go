package auth

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// MockUserManagementService provides a mock for testing auth commands.
type MockUserManagementService struct {
	ListUsersErr    error
	ListUsersResult []*megaport.User
}

func (m *MockUserManagementService) ListCompanyUsers(_ context.Context) ([]*megaport.User, error) {
	if m.ListUsersErr != nil {
		return nil, m.ListUsersErr
	}
	if m.ListUsersResult != nil {
		return m.ListUsersResult, nil
	}
	return []*megaport.User{}, nil
}

// Stubs for the full UserManagementService interface
func (m *MockUserManagementService) CreateUser(_ context.Context, _ *megaport.CreateUserRequest) (*megaport.CreateUserResponse, error) {
	return nil, nil
}
func (m *MockUserManagementService) GetUser(_ context.Context, _ int) (*megaport.User, error) {
	return nil, nil
}
func (m *MockUserManagementService) UpdateUser(_ context.Context, _ int, _ *megaport.UpdateUserRequest) error {
	return nil
}
func (m *MockUserManagementService) DeleteUser(_ context.Context, _ int) error {
	return nil
}
func (m *MockUserManagementService) DeactivateUser(_ context.Context, _ int) error {
	return nil
}
func (m *MockUserManagementService) GetUserActivity(_ context.Context, _ *megaport.GetUserActivityRequest) ([]*megaport.UserActivity, error) {
	return nil, nil
}
