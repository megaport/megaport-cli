package auth

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

var listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
	return client.UserManagementService.ListCompanyUsers(ctx)
}
