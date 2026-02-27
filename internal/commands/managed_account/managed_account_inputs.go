package managed_account

import (
	"encoding/json"
	"fmt"
	"os"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func parseManagedAccountRequestJSON(jsonStr, jsonFile string) (*megaport.ManagedAccountRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
	}

	req := &megaport.ManagedAccountRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return req, nil
}

func buildManagedAccountRequestFromFlags(cmd *cobra.Command) (*megaport.ManagedAccountRequest, error) { //nolint:unparam
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
