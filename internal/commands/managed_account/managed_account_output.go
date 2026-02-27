package managed_account

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// ManagedAccountOutput represents the desired fields for output of managed account details.
type ManagedAccountOutput struct {
	output.Output `json:"-" header:"-"`
	AccountName   string `json:"account_name" header:"ACCOUNT NAME"`
	AccountRef    string `json:"account_ref" header:"ACCOUNT REF"`
	CompanyUID    string `json:"company_uid" header:"COMPANY UID"`
}

// ToManagedAccountOutput converts a *megaport.ManagedAccount to our ManagedAccountOutput struct.
func ToManagedAccountOutput(account *megaport.ManagedAccount) (ManagedAccountOutput, error) {
	if account == nil {
		return ManagedAccountOutput{}, fmt.Errorf("invalid managed account: nil value")
	}

	return ManagedAccountOutput{
		AccountName: account.AccountName,
		AccountRef:  account.AccountRef,
		CompanyUID:  account.CompanyUID,
	}, nil
}

// printManagedAccounts prints a list of managed accounts in the specified format.
func printManagedAccounts(accounts []*megaport.ManagedAccount, format string, noColor bool) error {
	outputs := make([]ManagedAccountOutput, 0, len(accounts))
	for _, account := range accounts {
		o, err := ToManagedAccountOutput(account)
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

	fmt.Println()
	output.PrintInfo("Changes applied:", noColor)

	changesFound := false

	if original.AccountName != updated.AccountName {
		changesFound = true
		fmt.Printf("  • Account Name: %s → %s\n",
			output.FormatOldValue(original.AccountName, noColor),
			output.FormatNewValue(updated.AccountName, noColor))
	}

	if original.AccountRef != updated.AccountRef {
		changesFound = true
		fmt.Printf("  • Account Ref: %s → %s\n",
			output.FormatOldValue(original.AccountRef, noColor),
			output.FormatNewValue(updated.AccountRef, noColor))
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}
