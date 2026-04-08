package managed_account

import (
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func parseManagedAccountRequestJSON(jsonStr, jsonFile string) (*megaport.ManagedAccountRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.ManagedAccountRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return req, nil
}

func buildManagedAccountRequestFromFlags(cmd *cobra.Command) (*megaport.ManagedAccountRequest, error) { //nolint:unparam
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	accountName, _ := cmd.Flags().GetString("account-name")
	accountRef, _ := cmd.Flags().GetString("account-ref")

	req := &megaport.ManagedAccountRequest{
		AccountName: accountName,
		AccountRef:  accountRef,
	}

	return req, nil
}

func buildManagedAccountRequestFromJSON(jsonStr, jsonFile string) (*megaport.ManagedAccountRequest, error) {
	return parseManagedAccountRequestJSON(jsonStr, jsonFile)
}

func buildUpdateManagedAccountRequestFromFlags(cmd *cobra.Command) (*megaport.ManagedAccountRequest, error) { //nolint:unparam
	req := &megaport.ManagedAccountRequest{}

	if cmd.Flags().Changed("account-name") {
		accountName, _ := cmd.Flags().GetString("account-name")
		req.AccountName = accountName
	}

	if cmd.Flags().Changed("account-ref") {
		accountRef, _ := cmd.Flags().GetString("account-ref")
		req.AccountRef = accountRef
	}

	return req, nil
}

func buildUpdateManagedAccountRequestFromJSON(jsonStr, jsonFile string) (*megaport.ManagedAccountRequest, error) {
	return parseManagedAccountRequestJSON(jsonStr, jsonFile)
}
