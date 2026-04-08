package managed_account

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// managedAccountOutput represents the desired fields for output of managed account details.
type managedAccountOutput struct {
	output.Output `json:"-" header:"-"`
	AccountName   string `json:"account_name" header:"Account Name"`
	AccountRef    string `json:"account_ref" header:"Account Ref"`
	CompanyUID    string `json:"company_uid" header:"Company UID"`
}

// toManagedAccountOutput converts a *megaport.ManagedAccount to our managedAccountOutput struct.
func toManagedAccountOutput(account *megaport.ManagedAccount) (managedAccountOutput, error) {
	if account == nil {
		return managedAccountOutput{}, fmt.Errorf("invalid managed account: nil value")
	}

	return managedAccountOutput{
		AccountName: account.AccountName,
		AccountRef:  account.AccountRef,
		CompanyUID:  account.CompanyUID,
	}, nil
}

// printManagedAccounts prints a list of managed accounts in the specified format.
func printManagedAccounts(accounts []*megaport.ManagedAccount, format string, noColor bool) error {
	outputs := make([]managedAccountOutput, 0, len(accounts))
	for _, account := range accounts {
		o, err := toManagedAccountOutput(account)
		if err != nil {
			return err
		}
		outputs = append(outputs, o)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayManagedAccountChanges compares the original and updated managed account and displays the differences.
func displayManagedAccountChanges(original, updated *megaport.ManagedAccount, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	changes := []output.FieldChange{
		{Label: "Account Name", OldValue: original.AccountName, NewValue: updated.AccountName},
		{Label: "Account Ref", OldValue: original.AccountRef, NewValue: updated.AccountRef},
	}
	output.DisplayChanges(changes, noColor)
}
