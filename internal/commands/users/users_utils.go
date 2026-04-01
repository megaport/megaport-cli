package users

import (
	"context"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

var getUserFunc = func(ctx context.Context, client *megaport.Client, employeeID int) (*megaport.User, error) {
	return client.UserManagementService.GetUser(ctx, employeeID)
}

var createUserFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateUserRequest) (*megaport.CreateUserResponse, error) {
	return client.UserManagementService.CreateUser(ctx, req)
}

var updateUserFunc = func(ctx context.Context, client *megaport.Client, employeeID int, req *megaport.UpdateUserRequest) error {
	return client.UserManagementService.UpdateUser(ctx, employeeID, req)
}

var deleteUserFunc = func(ctx context.Context, client *megaport.Client, employeeID int) error {
	return client.UserManagementService.DeleteUser(ctx, employeeID)
}

var deactivateUserFunc = func(ctx context.Context, client *megaport.Client, employeeID int) error {
	return client.UserManagementService.DeactivateUser(ctx, employeeID)
}

var listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
	return client.UserManagementService.ListCompanyUsers(ctx)
}

var getUserActivityFunc = func(ctx context.Context, client *megaport.Client, req *megaport.GetUserActivityRequest) ([]*megaport.UserActivity, error) {
	return client.UserManagementService.GetUserActivity(ctx, req)
}

func filterUsers(users []*megaport.User, position string, activeOnly, inactiveOnly bool) []*megaport.User {
	var filtered []*megaport.User
	if users == nil {
		return filtered
	}
	for _, user := range users {
		if user == nil {
			continue
		}
		if activeOnly && !user.Active {
			continue
		}
		if inactiveOnly && user.Active {
			continue
		}
		if position != "" && !strings.EqualFold(user.Position, position) {
			continue
		}
		filtered = append(filtered, user)
	}
	return filtered
}
