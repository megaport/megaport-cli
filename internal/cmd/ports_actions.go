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
		PrintInfo("Using JSON input")
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input")
		req, err = processFlagPortInput(cmd)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode")
		req, err = promptForPortDetails()
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Buying port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to buy port: %v", err)
		return err
	}

	PrintResourceCreated("Port", resp.TechnicalServiceUIDs[0])
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
		PrintInfo("Using JSON input")
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input")
		req, err = processFlagLAGPortInput(cmd)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode")
		req, err = promptForLAGPortDetails()
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Buying LAG port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to buy LAG port: %v", err)
		return err
	}

	PrintResourceCreated("LAG Port", resp.TechnicalServiceUIDs[0])
	return nil
}

func ListPorts(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into Megaport API
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get all ports
	PrintInfo("Retrieving ports...")
	ports, err := client.PortService.ListPorts(ctx)
	if err != nil {
		PrintError("Failed to list ports: %v", err)
		return fmt.Errorf("error listing ports: %v", err)
	}

	// Get filter values from flags
	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")

	// Apply filters
	filteredPorts := filterPorts(ports, locationID, portSpeed, portName)

	if len(filteredPorts) == 0 {
		PrintWarning("No ports found matching the specified filters")
	}

	// Print ports with current output format
	err = printPorts(filteredPorts, outputFormat)
	if err != nil {
		PrintError("Failed to print ports: %v", err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func GetPort(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

	// Retrieve port details using the API client.
	PrintInfo("Retrieving port %s...", formattedUID)
	port, err := client.PortService.GetPort(ctx, portUID)
	if err != nil {
		PrintError("Failed to get port: %v", err)
		return fmt.Errorf("error getting port: %v", err)
	}

	if port == nil {
		PrintError("No port found with UID: %s", portUID)
		return fmt.Errorf("no port found with UID: %s", portUID)
	}

	// Print the port details using the desired output format.
	err = printPorts([]*megaport.Port{port}, outputFormat)
	if err != nil {
		PrintError("Failed to print ports: %v", err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func UpdatePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

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
		PrintInfo("Using JSON input for port %s", formattedUID)
		req, err = processJSONUpdatePortInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
		// Make sure the PortID from the command line arguments is set
		req.PortID = portUID
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input for port %s", formattedUID)
		req, err = processFlagUpdatePortInput(cmd, portUID)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode for port %s", formattedUID)
		req, err = promptForUpdatePortDetails(portUID)
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyPort method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Updating port %s...", formattedUID)
	resp, err := updatePortFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to update port: %v", err)
		return err
	}

	if resp.IsUpdated {
		PrintResourceUpdated("Port", portUID)
	} else {
		PrintWarning("Port update request was not successful")
	}
	return nil
}

func DeletePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

	// Get delete now flag
	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		PrintError("Failed to get delete now flag: %v", err)
		return err
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		PrintError("Failed to get force flag: %v", err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete port " + portUID + "? "
		if !confirmPrompt(confirmMsg) {
			PrintInfo("Deletion cancelled")
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
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Deleting port %s...", formattedUID)
	resp, err := deletePortFunc(ctx, client, deleteRequest)
	if err != nil {
		PrintError("Failed to delete port: %v", err)
		return err
	}

	if resp.IsDeleting {
		PrintResourceDeleted("Port", portUID, deleteNow)
	} else {
		PrintWarning("Port deletion request was not successful")
	}
	return nil
}

func RestorePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

	// Restore the port
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Restoring port %s...", formattedUID)
	resp, err := restorePortFunc(ctx, client, portUID)
	if err != nil {
		PrintError("Failed to restore port: %v", err)
		return err
	}

	if resp.IsRestored {
		PrintInfo("Port %s restored successfully", formattedUID)
	} else {
		PrintWarning("Port restoration request was not successful")
	}
	return nil
}

func LockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

	// Lock the port
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Locking port %s...", formattedUID)
	resp, err := lockPortFunc(ctx, client, portUID)
	if err != nil {
		PrintError("Failed to lock port: %v", err)
		return err
	}

	if resp.IsLocking {
		PrintInfo("Port %s locked successfully", formattedUID)
	} else {
		PrintWarning("Port lock request was not successful")
	}
	return nil
}

func UnlockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := formatUID(portUID)

	// Unlock the port
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Unlocking port %s...", formattedUID)
	resp, err := unlockPortFunc(ctx, client, portUID)
	if err != nil {
		PrintError("Failed to unlock port: %v", err)
		return err
	}

	if resp.IsUnlocking {
		PrintInfo("Port %s unlocked successfully", formattedUID)
	} else {
		PrintWarning("Port unlock request was not successful")
	}
	return nil
}

func CheckPortVLANAvailability(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID and VLAN ID from the command line arguments.
	portUID := args[0]
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		PrintError("Invalid VLAN ID: %v", err)
		return fmt.Errorf("invalid VLAN ID")
	}
	formattedUID := formatUID(portUID)

	// Check VLAN availability
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Checking VLAN %d availability on port %s...", vlan, formattedUID)
	available, err := checkPortVLANAvailabilityFunc(ctx, client, portUID, vlan)
	if err != nil {
		PrintError("Failed to check VLAN availability: %v", err)
		return err
	}

	if available {
		PrintInfo("VLAN %d is available on port %s", vlan, formattedUID)
	} else {
		PrintWarning("VLAN %d is not available on port %s", vlan, formattedUID)
	}
	return nil
}
