package ports

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyPort(cmd *cobra.Command, args []string, noColor bool) error {
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
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = promptForPortDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	// Validate the Port Request
	output.PrintInfo("Validating port order...", noColor)
	err = client.PortService.ValidatePortOrder(ctx, req)
	if err != nil {
		output.PrintError("Failed to validate port request: %v", noColor, err)
		return err
	}

	// Start spinner for creating port
	spinner := output.PrintResourceCreating("Port", req.Name, noColor)

	resp, err := buyPortFunc(ctx, client, req)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy port: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("Port", resp.TechnicalServiceUIDs[0], noColor)
	return nil
}

func BuyLAGPort(cmd *cobra.Command, args []string, noColor bool) error {
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
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagLAGPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = promptForLAGPortDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
	}
	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyPort method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for creating LAG port
	spinner := output.PrintResourceCreating("LAG Port", req.Name, noColor)

	resp, err := buyPortFunc(ctx, client, req)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy LAG port: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("LAG Port", resp.TechnicalServiceUIDs[0], noColor)
	return nil
}

func ListPorts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into Megaport API
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Start spinner for listing ports
	spinner := output.PrintResourceListing("Port", noColor)

	// Get all ports
	ports, err := client.PortService.ListPorts(ctx)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list ports: %v", noColor, err)
		return fmt.Errorf("error listing ports: %v", err)
	}

	// Get filter values from flags
	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	// Apply filters
	filteredPorts := filterPorts(ports, locationID, portSpeed, portName, includeInactive)

	if len(filteredPorts) == 0 {
		output.PrintWarning("No ports found matching the specified filters", noColor)
	}

	// output.Print ports with current output format
	err = printPorts(filteredPorts, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func GetPort(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Start spinner for getting port details
	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	// Retrieve port details using the API client.
	port, err := client.PortService.GetPort(ctx, portUID)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get port: %v", noColor, err)
		return fmt.Errorf("error getting port: %v", err)
	}

	if port == nil {
		output.PrintError("No port found with UID: %s", noColor, portUID)
		return fmt.Errorf("no port found with UID: %s", portUID)
	}

	// output.Print the port details using the desired output format.
	err = printPorts([]*megaport.Port{port}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

// UpdatePort handles updating an existing port
func UpdatePort(cmd *cobra.Command, args []string, noColor bool) error {
	// Initialize context and get client
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return err
	}

	portUID := args[0]

	// Check which input mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	var req *megaport.ModifyPortRequest

	// Start spinner for getting original port details
	getSpinner := output.PrintResourceGetting("Port", portUID, noColor)

	// Retrieve the original port for comparison
	originalPort, err := getPortFunc(ctx, client, portUID)

	// Stop spinner
	getSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve original port: %v", noColor, err)
		return err
	}

	if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = promptForUpdatePortDetails(portUID, noColor)
	} else if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONUpdatePortInput(jsonStr, jsonFile)
		if err == nil { // Only set portID if there was no error
			req.PortID = portUID // Set the port ID from the args
		}
	} else if cmd.Flags().Changed("name") || cmd.Flags().Changed("marketplace-visibility") ||
		cmd.Flags().Changed("cost-centre") || cmd.Flags().Changed("term") {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagUpdatePortInput(cmd, portUID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	if err != nil {
		return fmt.Errorf("failed to process input: %v", err)
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Start spinner for updating port
	updateSpinner := output.PrintResourceUpdating("Port", portUID, noColor)

	// Call the API
	resp, err := updatePortFunc(ctx, client, req)

	// Stop spinner
	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to update port: %v", noColor, err)
		return err
	}

	// Check the response
	if !resp.IsUpdated {
		output.PrintError("Port update request was not successful", noColor)
		return fmt.Errorf("port update request was not successful")
	}

	output.PrintResourceUpdated("Port", portUID, noColor)

	// Start spinner for getting updated port details
	getUpdatedSpinner := output.PrintResourceGetting("Port", portUID, noColor)

	// Retrieve the updated port for comparison
	updatedPort, err := getPortFunc(ctx, client, portUID)

	// Stop spinner
	getUpdatedSpinner.Stop()

	if err != nil {
		output.PrintError("Port was updated but failed to retrieve updated details: %v", noColor, err)
		return nil
	}

	// Display changes between original and updated port
	displayPortChanges(originalPort, updatedPort, noColor)

	return nil
}

func DeletePort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Get delete now flag
	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		output.PrintError("Failed to get delete now flag: %v", noColor, err)
		return err
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		output.PrintError("Failed to get force flag: %v", noColor, err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete port " + portUID + "? "
		if !utils.ConfirmPrompt(confirmMsg, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeletePortRequest{
		PortID:    portUID,
		DeleteNow: deleteNow,
	}

	// Delete the port
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for deleting port
	spinner := output.PrintResourceDeleting("Port", portUID, noColor)

	resp, err := deletePortFunc(ctx, client, deleteRequest)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to delete port: %v", noColor, err)
		return err
	}

	if resp.IsDeleting {
		output.PrintResourceDeleted("Port", portUID, deleteNow, noColor)
	} else {
		output.PrintWarning("Port deletion request was not successful", noColor)
	}
	return nil
}

func RestorePort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	// Restore the port
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for restoring port
	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := restorePortFunc(ctx, client, portUID)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to restore port: %v", noColor, err)
		return err
	}

	if resp.IsRestored {
		output.PrintInfo("Port %s restored successfully", noColor, formattedUID)
	} else {
		output.PrintWarning("Port restoration request was not successful", noColor)
	}
	return nil
}

func LockPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	// Lock the port
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for locking port
	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := lockPortFunc(ctx, client, portUID)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to lock port: %v", noColor, err)
		return err
	}

	if resp.IsLocking {
		output.PrintInfo("Port %s locked successfully", noColor, formattedUID)
	} else {
		output.PrintWarning("Port lock request was not successful", noColor)
	}
	return nil
}

func UnlockPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	// Unlock the port
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for unlocking port
	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := unlockPortFunc(ctx, client, portUID)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to unlock port: %v", noColor, err)
		return err
	}

	if resp.IsUnlocking {
		output.PrintInfo("Port %s unlocked successfully", noColor, formattedUID)
	} else {
		output.PrintWarning("Port unlock request was not successful", noColor)
	}
	return nil
}

func CheckPortVLANAvailability(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Retrieve the port UID and VLAN ID from the command line arguments.
	portUID := args[0]
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		output.PrintError("Invalid VLAN ID: %v", noColor, err)
		return fmt.Errorf("invalid VLAN ID")
	}

	// Check VLAN availability
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for checking VLAN availability
	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	available, err := checkPortVLANAvailabilityFunc(ctx, client, portUID, vlan)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to check VLAN availability: %v", noColor, err)
		return err
	}

	if available {
		output.PrintInfo("VLAN %d is available on port %s", noColor, vlan, output.FormatUID(portUID, noColor))
	} else {
		output.PrintWarning("VLAN %d is not available on port %s", noColor, vlan, output.FormatUID(portUID, noColor))
	}
	return nil
}
