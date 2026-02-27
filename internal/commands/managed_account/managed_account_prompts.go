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

func buildUpdateManagedAccountRequestFromPrompt(noColor bool) (*megaport.ManagedAccountRequest, error) {
	req := &megaport.ManagedAccountRequest{}
	fieldsUpdated := false

	accountName, err := utils.ResourcePrompt("managed-account", "Enter new account name (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if accountName != "" {
		req.AccountName = accountName
		fieldsUpdated = true
	}

	accountRef, err := utils.ResourcePrompt("managed-account", "Enter new account reference (leave empty to skip): ", noColor)
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
