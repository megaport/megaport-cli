package managed_account

import (
	"context"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

var createManagedAccountFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
	return client.ManagedAccountService.CreateManagedAccount(ctx, req)
}

var updateManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
	return client.ManagedAccountService.UpdateManagedAccount(ctx, companyUID, req)
}

var getManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, name string) (*megaport.ManagedAccount, error) {
	return client.ManagedAccountService.GetManagedAccount(ctx, companyUID, name)
}

func filterManagedAccounts(accounts []*megaport.ManagedAccount, accountName, accountRef string) []*megaport.ManagedAccount {
	return utils.Filter(accounts, func(account *megaport.ManagedAccount) bool {
		if account == nil {
			return false
		}
		if accountName != "" && !strings.Contains(strings.ToLower(account.AccountName), strings.ToLower(accountName)) {
			return false
		}
		if accountRef != "" && !strings.Contains(strings.ToLower(account.AccountRef), strings.ToLower(accountRef)) {
			return false
		}
		return true
	})
}
