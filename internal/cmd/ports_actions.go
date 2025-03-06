package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("marketplace-visibility")

	var req *megaport.BuyPortRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagPortInput(cmd)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForPortDetails()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("Port purchased successfully - UID: %s\n", resp.TechnicalServiceUIDs[0])
	return nil
}
func BuyLAGPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("lag-count") || cmd.Flags().Changed("marketplace-visibility")

	var req *megaport.BuyPortRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagLAGPortInput(cmd)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForLAGPortDetails()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying LAG port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("LAG port purchased successfully - UID: %s\n", resp.TechnicalServiceUIDs[0])
	return nil
}

func ListPorts(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into Megaport API
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get all ports
	ports, err := client.PortService.ListPorts(ctx)
	if err != nil {
		return fmt.Errorf("error listing ports: %v", err)
	}

	// Get filter values from flags
	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")

	// Apply filters
	filteredPorts := filterPorts(ports, locationID, portSpeed, portName)

	// Print ports with current output format
	return printPorts(filteredPorts, outputFormat)
}

func GetPort(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Retrieve port details using the API client.
	port, err := client.PortService.GetPort(ctx, portUID)
	if err != nil {
		return fmt.Errorf("error getting port: %v", err)
	}

	// Print the port details using the desired output format.
	err = printPorts([]*megaport.Port{port}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func UpdatePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("marketplace-visibility") ||
		cmd.Flags().Changed("cost-centre") || cmd.Flags().Changed("term")

	var req *megaport.ModifyPortRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		req, err = processJSONUpdatePortInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
		// Make sure the PortID from the command line arguments is set
		req.PortID = portUID
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagUpdatePortInput(cmd, portUID)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForUpdatePortDetails(portUID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating port...")
	resp, err := updatePortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	if resp.IsUpdated {
		fmt.Printf("Port updated successfully - UID: %s\n", portUID)
	} else {
		fmt.Println("Port update request was not successful")
	}
	return nil
}

func DeletePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

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
		confirmMsg := "Are you sure you want to delete port " + portUID + "? (y/n): "
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
	deleteRequest := &megaport.DeletePortRequest{
		PortID:    portUID,
		DeleteNow: deleteNow,
	}

	// Delete the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := deletePortFunc(ctx, client, deleteRequest)
	if err != nil {
		return err
	}

	if resp.IsDeleting {
		fmt.Printf("Port %s deleted successfully\n", portUID)
		if deleteNow {
			fmt.Println("The port will be deleted immediately")
		} else {
			fmt.Println("The port will be deleted at the end of the current billing period")
		}
	} else {
		fmt.Println("Port deletion request was not successful")
	}
	return nil
}

func RestorePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Restore the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := restorePortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsRestored {
		fmt.Printf("Port %s restored successfully\n", portUID)
	} else {
		fmt.Println("Port restoration request was not successful")
	}
	return nil
}

func LockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Lock the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := lockPortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsLocking {
		fmt.Printf("Port %s locked successfully\n", portUID)
	} else {
		fmt.Println("Port lock request was not successful")
	}
	return nil
}

func UnlockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Unlock the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := unlockPortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsUnlocking {
		fmt.Printf("Port %s unlocked successfully\n", portUID)
	} else {
		fmt.Println("Port unlock request was not successful")
	}
	return nil
}

func CheckPortVLANAvailability(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID and VLAN ID from the command line arguments.
	portUID := args[0]
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid VLAN ID")
	}

	// Check VLAN availability
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	available, err := checkPortVLANAvailabilityFunc(ctx, client, portUID, vlan)
	if err != nil {
		return err
	}

	if available {
		fmt.Printf("VLAN %d is available on port %s\n", vlan, portUID)
	} else {
		fmt.Printf("VLAN %d is not available on port %s\n", vlan, portUID)
	}
	return nil
}
