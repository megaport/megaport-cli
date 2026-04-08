package managed_account

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func ListManagedAccounts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	accountName, _ := cmd.Flags().GetString("account-name")
	accountRef, _ := cmd.Flags().GetString("account-ref")

	spinner := output.PrintResourceListing("managed account", noColor)

	accounts, err := client.ManagedAccountService.ListManagedAccounts(ctx)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list managed accounts: %v", noColor, err)
		return fmt.Errorf("error listing managed accounts: %w", err)
	}

	filteredAccounts := filterManagedAccounts(accounts, accountName, accountRef)

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(filteredAccounts) > limit {
		filteredAccounts = filteredAccounts[:limit]
	}

	if len(filteredAccounts) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No managed accounts found.", noColor)
		}
		return nil
	}

	err = printManagedAccounts(filteredAccounts, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print managed accounts: %v", noColor, err)
		return fmt.Errorf("error printing managed accounts: %w", err)
	}
	return nil
}

func GetManagedAccount(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	companyUID := args[0]
	accountName := args[1]

	spinner := output.PrintResourceGetting("managed account", accountName, noColor)

	account, err := getManagedAccountFunc(ctx, client, companyUID, accountName)

	spinner.Stop()

	if err != nil {
		output.PrintError("Error getting managed account: %v", noColor, err)
		return fmt.Errorf("error getting managed account: %w", err)
	}

	err = printManagedAccounts([]*megaport.ManagedAccount{account}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing managed accounts: %w", err)
	}
	return nil
}

func CreateManagedAccount(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("account-name") || cmd.Flags().Changed("account-ref")

	var req *megaport.ManagedAccountRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildManagedAccountRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = buildManagedAccountRequestFromFlags(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = buildManagedAccountRequestFromPrompt(noColor)
		if err != nil {
			return err
		}
	} else {
		output.PrintError("No input provided, use --interactive, --json, or flags to specify managed account details", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify managed account details")
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceCreating("managed account", req.AccountName, noColor)
	account, err := createManagedAccountFunc(ctx, client, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Error creating managed account: %v", noColor, err)
		return err
	}

	output.PrintSuccess("Managed account created successfully - Company UID: %s", noColor, account.CompanyUID)
	return nil
}

func UpdateManagedAccount(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	companyUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("account-name") || cmd.Flags().Changed("account-ref")

	var req *megaport.ManagedAccountRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildUpdateManagedAccountRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = buildUpdateManagedAccountRequestFromFlags(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = buildUpdateManagedAccountRequestFromPrompt(noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}

	// Fetch original for change display
	accounts, listErr := client.ManagedAccountService.ListManagedAccounts(ctx)
	var originalAccount *megaport.ManagedAccount
	if listErr == nil {
		for _, a := range accounts {
			if a != nil && a.CompanyUID == companyUID {
				originalAccount = a
				break
			}
		}
	}

	updateSpinner := output.PrintResourceUpdating("managed account", companyUID, noColor)
	updatedAccount, err := updateManagedAccountFunc(ctx, client, companyUID, req)
	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Error updating managed account: %v", noColor, err)
		return err
	}

	output.PrintResourceUpdated("managed account", companyUID, noColor)

	displayManagedAccountChanges(originalAccount, updatedAccount, noColor)

	return nil
}
