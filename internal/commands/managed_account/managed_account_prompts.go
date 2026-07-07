package managed_account

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func buildManagedAccountRequestFromPrompt(noColor bool) (*megaport.ManagedAccountRequest, error) {
	accountName, err := utils.ResourcePrompt("managed-account", "Enter account name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if accountName == "" {
		return nil, fmt.Errorf("account name is required")
	}

	accountRef, err := utils.ResourcePrompt("managed-account", "Enter account reference (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if accountRef == "" {
		return nil, fmt.Errorf("account reference is required")
	}

	req := &megaport.ManagedAccountRequest{
		AccountName: accountName,
		AccountRef:  accountRef,
	}

	return req, nil
}

// buildUpdateManagedAccountRequestFromPrompt seeds the request from the current
// account, so a prompt the user leaves empty keeps its current value instead of
// being sent as an empty string.
func buildUpdateManagedAccountRequestFromPrompt(noColor bool, current *megaport.ManagedAccount) (*megaport.ManagedAccountRequest, error) {
	req := &megaport.ManagedAccountRequest{
		AccountName: current.AccountName,
		AccountRef:  current.AccountRef,
	}
	fieldsUpdated := false

	accountName, err := utils.ResourcePrompt("managed-account", "Enter new account name (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if accountName != "" {
		req.AccountName = accountName
		fieldsUpdated = true
	}

	accountRef, err := utils.ResourcePrompt("managed-account", "Enter new account reference (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if accountRef != "" {
		req.AccountRef = accountRef
		fieldsUpdated = true
	}

	if !fieldsUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}
