package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// mcrCmd is the parent command for all operations related to Megaport Cloud Routers (MCRs).
// It serves as a container for subcommands that manage and retrieve information about MCRs.
//
// Example usage:
//
//	megaport mcr get [mcrUID]
var mcrCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
	Long: `Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
You can use the subcommands to perform actions such as retrieving details for a specific MCR.
For instance, use the "megaport mcr get [mcrUID]" command to fetch details for the MCR with the given UID.
`,
}

// getMCRCmd retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
// This command requires exactly one argument: the UID of the MCR.
//
// It establishes a context with a timeout, logs into the Megaport API, and then uses the API client
// to get the MCR details. The retrieved information is printed using the configured output format (table/json).
//
// Example usage:
//
//	megaport mcr get MCR12345
var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		mcr, err := client.MCRService.GetMCR(ctx, mcrUID)
		if err != nil {
			return fmt.Errorf("error getting MCR: %v", err)
		}

		// Print the MCR details using the desired output format.
		err = printMCRs([]*megaport.MCR{mcr}, outputFormat)
		if err != nil {
			return fmt.Errorf("error printing MCRs: %v", err)
		}
		return nil
	},
}

var buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	// The real implementation will call mcrService.BuyMCR
	return client.MCRService.BuyMCR(ctx, req)
}

var buyMCRCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy an MCR through the Megaport API",
	Long: `Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the MCR.
  - term: The term of the MCR (1, 12, 24, or 36 months).
  - port_speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
  - location_id: The ID of the location where the MCR will be provisioned.

Optional fields:
  - diversity_zone: The diversity zone for the MCR.
  - cost_center: The cost center for the MCR.
  - promo_code: A promotional code for the MCR.

Example usage:

  megaport mcr buy
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// deleteMCRCmd deletes a Megaport Cloud Router (MCR) from the user's account.
// This command requires the MCR UID as an argument and will prompt for confirmation
// before proceeding with deletion unless the --force flag is used.
//
// Example usage:
//
//	megaport mcr delete MCR12345
var deleteMCRCmd = &cobra.Command{
	Use:   "delete [mcrUID]",
	Short: "Delete an MCR from your account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		resp, err := client.MCRService.DeleteMCR(ctx, deleteRequest)
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
	},
}

// restoreMCRCmd restores a previously deleted Megaport Cloud Router (MCR).
// This command requires the MCR UID as an argument.
//
// Example usage:
//
//	megaport mcr restore MCR12345
var restoreMCRCmd = &cobra.Command{
	Use:   "restore [mcrUID]",
	Short: "Restore a deleted MCR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		resp, err := client.MCRService.RestoreMCR(ctx, mcrUID)
		if err != nil {
			return fmt.Errorf("error restoring MCR: %v", err)
		}

		if resp.IsRestored {
			fmt.Printf("MCR %s restored successfully\n", mcrUID)
		} else {
			fmt.Println("MCR restoration request was not successful")
		}

		return nil
	},
}

func init() {
	mcrCmd.AddCommand(getMCRCmd)
	mcrCmd.AddCommand(buyMCRCmd)

	// Add delete command with flags
	deleteMCRCmd.Flags().Bool("now", false, "Delete immediately instead of at the end of the billing period")
	deleteMCRCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	mcrCmd.AddCommand(deleteMCRCmd)

	// Add restore command
	mcrCmd.AddCommand(restoreMCRCmd)

	rootCmd.AddCommand(mcrCmd)
}

type MCROutput struct {
	output
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	ProvisioningStatus string `json:"provisioning_status"`
}

func ToMCROutput(mcr *megaport.MCR) (MCROutput, error) {
	if mcr == nil {
		return MCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}

	return MCROutput{
		UID:                mcr.UID,
		Name:               mcr.Name,
		LocationID:         mcr.LocationID,
		ProvisioningStatus: mcr.ProvisioningStatus,
	}, nil
}

func printMCRs(mcrs []*megaport.MCR, format string) error {
	outputs := make([]MCROutput, 0, len(mcrs))
	for _, mcr := range mcrs {
		output, err := ToMCROutput(mcr)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}
