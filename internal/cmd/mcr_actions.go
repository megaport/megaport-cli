package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials.
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments.
	mcrUID := args[0]

	// Use the API client to get the MCR details based on the provided UID.
	mcr, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error getting MCR: %v", err)
	}

	// Print the MCR details using the desired output format.
	err = printMCRs([]*megaport.MCR{mcr}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func BuyMCR(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	name, err := prompt("Enter MCR name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("MCR name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	portSpeedStr, err := prompt("Enter port speed (1000, 2500, 5000, 10000) (required): ")
	if err != nil {
		return err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 1000 && portSpeed != 2500 && portSpeed != 5000 && portSpeed != 10000) {
		return fmt.Errorf("invalid port speed, must be one of 1000, 2500, 5000, 10000")
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return fmt.Errorf("invalid location ID")
	}

	// Prompt for optional fields
	diversityZone, err := prompt("Enter diversity zone (optional): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter cost center (optional): ")
	if err != nil {
		return err
	}

	promoCode, err := prompt("Enter promo code (optional): ")
	if err != nil {
		return err
	}

	// Create the BuyMCRRequest
	req := &megaport.BuyMCRRequest{
		Name:             name,
		Term:             term,
		PortSpeed:        portSpeed,
		LocationID:       locationID,
		DiversityZone:    diversityZone,
		CostCentre:       costCentre,
		PromoCode:        promoCode,
		WaitForProvision: true,
		WaitForTime:      10 * time.Minute,
	}

	// Call the BuyMCR method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying MCR...")

	if err := client.MCRService.ValidateMCROrder(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	resp, err := buyMCRFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("MCR purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

func UpdateMCR(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	mcrUID := args[0]

	// Prompt for fields to update
	name, err := prompt("Enter new MCR name (leave blank to keep current): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter new cost centre (leave blank to keep current): ")
	if err != nil {
		return err
	}

	marketplaceVisibilityStr, err := prompt("Enter new marketplace visibility (true/false, leave blank to keep current): ")
	if err != nil {
		return err
	}
	var marketplaceVisibility *bool
	if marketplaceVisibilityStr != "" {
		visibility, err := strconv.ParseBool(marketplaceVisibilityStr)
		if err != nil {
			return fmt.Errorf("invalid marketplace visibility, must be true or false")
		}
		marketplaceVisibility = &visibility
	}

	contractTermMonthsStr, err := prompt("Enter new contract term in months (leave blank to keep current): ")
	if err != nil {
		return err
	}
	var contractTermMonths *int
	if contractTermMonthsStr != "" {
		term, err := strconv.Atoi(contractTermMonthsStr)
		if err != nil {
			return fmt.Errorf("invalid contract term, must be a number")
		}
		contractTermMonths = &term
	}

	// Create the ModifyMCRRequest
	req := &megaport.ModifyMCRRequest{
		MCRID:                 mcrUID,
		Name:                  name,
		CostCentre:            costCentre,
		MarketplaceVisibility: marketplaceVisibility,
		ContractTermMonths:    contractTermMonths,
		WaitForUpdate:         true,
		WaitForTime:           10 * time.Minute,
	}

	// Call the ModifyMCR method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating MCR...")
	resp, err := client.MCRService.ModifyMCR(ctx, req)
	if err != nil {
		return err
	}

	if resp.IsUpdated {
		fmt.Println("MCR updated successfully")
	} else {
		fmt.Println("MCR update failed")
	}
	return nil
}

func DeleteMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Get delete now flag
	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		return err
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete MCR " + mcrUID + "? (y/n): "
		confirmation, err := prompt(confirmMsg)
		if err != nil {
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:     mcrUID,
		DeleteNow: deleteNow,
	}

	// Delete the MCR
	resp, err := deleteMCRFunc(ctx, client, deleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting MCR: %v", err)
	}

	if resp.IsDeleting {
		fmt.Printf("MCR %s deleted successfully\n", mcrUID)
		if deleteNow {
			fmt.Println("The MCR will be deleted immediately")
		} else {
			fmt.Println("The MCR will be deleted at the end of the current billing period")
		}
	} else {
		fmt.Println("MCR deletion request was not successful")
	}

	return nil
}

func RestoreMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Restore the MCR
	resp, err := restoreMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error restoring MCR: %v", err)
	}

	if resp.IsRestored {
		fmt.Printf("MCR %s restored successfully\n", mcrUID)
	} else {
		fmt.Println("MCR restoration request was not successful")
	}

	return nil
}

func CreateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Prompt for required fields
	description, err := prompt("Enter prefix filter list description (required): ")
	if err != nil {
		return err
	}
	if description == "" {
		return fmt.Errorf("description is required")
	}

	addressFamily, err := prompt("Enter address family (IPv4 or IPv6) (required): ")
	if err != nil {
		return err
	}
	if addressFamily != "IPv4" && addressFamily != "IPv6" {
		return fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}

	// Prompt for prefix filter list entries
	var entries []*megaport.MCRPrefixListEntry
	for {
		action, err := prompt("Enter action (permit or deny) (required, leave empty to finish): ")
		if err != nil {
			return err
		}
		if action == "" {
			break
		}
		if action != "permit" && action != "deny" {
			return fmt.Errorf("invalid action, must be permit or deny")
		}

		prefix, err := prompt("Enter prefix (required): ")
		if err != nil {
			return err
		}
		if prefix == "" {
			return fmt.Errorf("prefix is required")
		}

		geStr, err := prompt("Enter greater than or equal to (optional): ")
		if err != nil {
			return err
		}
		var ge int
		if geStr != "" {
			ge, err = strconv.Atoi(geStr)
			if err != nil {
				return fmt.Errorf("invalid greater than or equal to value")
			}
		}

		leStr, err := prompt("Enter less than or equal to (optional): ")
		if err != nil {
			return err
		}
		var le int
		if leStr != "" {
			le, err = strconv.Atoi(leStr)
			if err != nil {
				return fmt.Errorf("invalid less than or equal to value")
			}
		}

		entry := &megaport.MCRPrefixListEntry{
			Action: action,
			Prefix: prefix,
			Ge:     ge,
			Le:     le,
		}
		entries = append(entries, entry)
	}

	// Create the CreateMCRPrefixFilterListRequest
	req := &megaport.CreateMCRPrefixFilterListRequest{
		MCRID: mcrUID,
		PrefixFilterList: megaport.MCRPrefixFilterList{
			Description:   description,
			AddressFamily: addressFamily,
			Entries:       entries,
		},
	}

	// Call the CreatePrefixFilterList method
	resp, err := createMCRPrefixFilterListFunc(ctx, client, req)
	if err != nil {
		return fmt.Errorf("error creating prefix filter list: %v", err)
	}

	fmt.Printf("Prefix filter list created successfully - ID: %d\n", resp.PrefixFilterListID)
	return nil
}

func ListMCRPrefixFilterLists(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Call the ListMCRPrefixFilterLists method
	prefixFilterLists, err := listMCRPrefixFilterListsFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error listing prefix filter lists: %v", err)
	}

	// Print the prefix filter lists using the desired output format
	err = printOutput(prefixFilterLists, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing prefix filter lists: %v", err)
	}
	return nil
}

func GetMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Call the GetMCRPrefixFilterList method
	prefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return fmt.Errorf("error getting prefix filter list: %v", err)
	}

	// Convert the prefix filter list to the custom output format
	output, err := ToPrefixFilterListOutput(prefixFilterList)
	if err != nil {
		return fmt.Errorf("error converting prefix filter list: %v", err)
	}

	// Print the prefix filter list details using the desired output format
	err = printOutput([]PrefixFilterListOutput{output}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing prefix filter list: %v", err)
	}
	return nil
}

func UpdateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Prompt for required fields
	description, err := prompt("Enter new prefix filter list description (required): ")
	if err != nil {
		return err
	}
	if description == "" {
		return fmt.Errorf("description is required")
	}

	addressFamily, err := prompt("Enter new address family (IPv4 or IPv6) (required): ")
	if err != nil {
		return err
	}
	if addressFamily != "IPv4" && addressFamily != "IPv6" {
		return fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}

	// Prompt for prefix filter list entries
	var entries []*megaport.MCRPrefixListEntry
	for {
		action, err := prompt("Enter action (permit or deny) (required, leave empty to finish): ")
		if err != nil {
			return err
		}
		if action == "" {
			break
		}
		if action != "permit" && action != "deny" {
			return fmt.Errorf("invalid action, must be permit or deny")
		}

		prefix, err := prompt("Enter prefix (required): ")
		if err != nil {
			return err
		}
		if prefix == "" {
			return fmt.Errorf("prefix is required")
		}

		geStr, err := prompt("Enter greater than or equal to (optional): ")
		if err != nil {
			return err
		}
		var ge int
		if geStr != "" {
			ge, err = strconv.Atoi(geStr)
			if err != nil {
				return fmt.Errorf("invalid greater than or equal to value")
			}
		}

		leStr, err := prompt("Enter less than or equal to (optional): ")
		if err != nil {
			return err
		}
		var le int
		if leStr != "" {
			le, err = strconv.Atoi(leStr)
			if err != nil {
				return fmt.Errorf("invalid less than or equal to value")
			}
		}

		entry := &megaport.MCRPrefixListEntry{
			Action: action,
			Prefix: prefix,
			Ge:     ge,
			Le:     le,
		}
		entries = append(entries, entry)
	}

	// Create the ModifyMCRPrefixFilterListRequest
	req := &megaport.MCRPrefixFilterList{
		Description:   description,
		AddressFamily: addressFamily,
		Entries:       entries,
	}
	// Call the ModifyMCRPrefixFilterList method
	resp, err := modifyMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID, req)
	if err != nil {
		return fmt.Errorf("error updating prefix filter list: %v", err)
	}

	if resp.IsUpdated {
		fmt.Printf("Prefix filter list updated successfully - ID: %d\n", prefixFilterListID)
	} else {
		fmt.Println("Prefix filter list update request was not successful")
	}

	return nil
}

func DeleteMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Call the DeleteMCRPrefixFilterList method
	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return fmt.Errorf("error deleting prefix filter list: %v", err)
	}

	if resp.IsDeleted {
		fmt.Printf("Prefix filter list deleted successfully - ID: %d\n", prefixFilterListID)
	} else {
		fmt.Println("Prefix filter list deletion request was not successful")
	}

	return nil
}
