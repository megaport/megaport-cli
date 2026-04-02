package ports

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func buildPortRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyPortRequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("marketplace-visibility")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := processFlagPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err := promptForPortDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
}

func buildLAGPortRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyPortRequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("lag-count") || cmd.Flags().Changed("marketplace-visibility")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := processFlagLAGPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err := promptForLAGPortDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify port details")
}

func BuyPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = 10 * time.Minute
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	validateSpinner := output.PrintResourceValidating("Port", noColor)
	err = client.PortService.ValidatePortOrder(ctx, req)
	validateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate port request: %v", noColor, err)
		return err
	}

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.Name},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Port Speed", Value: fmt.Sprintf("%d Mbps", req.PortSpeed)},
			{Key: "Location ID", Value: strconv.Itoa(req.LocationId)},
		}
		if !utils.BuyConfirmPrompt("Port", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	var spinner *output.Spinner
	if req.WaitForProvision {
		spinner = output.PrintResourceProvisioning("Port", req.Name, noColor)
	} else {
		spinner = output.PrintResourceCreating("Port", req.Name, noColor)
	}

	resp, err := buyPortFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy port: %v", noColor, err)
		return err
	}

	if len(resp.TechnicalServiceUIDs) == 0 {
		output.PrintError("Port created but no UID returned", noColor)
		return fmt.Errorf("port created but no UID returned")
	}

	output.PrintResourceCreated("Port", resp.TechnicalServiceUIDs[0], noColor)
	return nil
}

func ValidatePort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceValidating("Port", noColor)
	err = client.PortService.ValidatePortOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate port request: %v", noColor, err)
		return err
	}

	output.PrintSuccess("Port validation passed", noColor)
	return nil
}

func ValidateLAGPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildLAGPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceValidating("LAG Port", noColor)
	err = client.PortService.ValidatePortOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate LAG port request: %v", noColor, err)
		return err
	}

	output.PrintSuccess("LAG Port validation passed", noColor)
	return nil
}

func BuyLAGPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildLAGPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = 10 * time.Minute
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	validateSpinner := output.PrintResourceValidating("LAG Port", noColor)
	err = client.PortService.ValidatePortOrder(ctx, req)
	validateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate LAG port request: %v", noColor, err)
		return err
	}

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.Name},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Port Speed", Value: fmt.Sprintf("%d Mbps", req.PortSpeed)},
			{Key: "Location ID", Value: strconv.Itoa(req.LocationId)},
			{Key: "LAG Count", Value: strconv.Itoa(req.LagCount)},
		}
		if !utils.BuyConfirmPrompt("LAG Port", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	var spinner *output.Spinner
	if req.WaitForProvision {
		spinner = output.PrintResourceProvisioning("LAG Port", req.Name, noColor)
	} else {
		spinner = output.PrintResourceCreating("LAG Port", req.Name, noColor)
	}

	resp, err := buyPortFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy LAG port: %v", noColor, err)
		return err
	}

	if len(resp.TechnicalServiceUIDs) == 0 {
		output.PrintError("LAG port created but no UID returned", noColor)
		return fmt.Errorf("LAG port created but no UID returned")
	}

	output.PrintResourceCreated("LAG Port", resp.TechnicalServiceUIDs[0], noColor)
	return nil
}

// listPortsFunc is a variable that can be overridden by WASM builds
var listPortsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Port, error) {
	return client.PortService.ListPorts(ctx)
}

func ListPorts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("Port", noColor)

	ports, err := listPortsFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list ports: %v", noColor, err)
		return fmt.Errorf("error listing ports: %v", err)
	}

	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	filteredPorts := filterPorts(ports, locationID, portSpeed, portName, includeInactive)

	if len(filteredPorts) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No ports found. Create one with 'megaport ports buy'.", noColor)
		}
		return nil
	}

	err = printPorts(filteredPorts, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func GetPort(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := getPortFunc(ctx, client, portUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "Port", portUID)
		output.PrintError("Failed to get port: %v", noColor, err)
		return fmt.Errorf("error getting port: %w", err)
	}

	if port == nil {
		output.PrintError("No port found with UID: %s", noColor, portUID)
		return fmt.Errorf("no port found with UID: %s", portUID)
	}

	err = printPorts([]*megaport.Port{port}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func GetPortStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := getPortFunc(ctx, client, portUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "Port", portUID)
		output.PrintError("Failed to get Port status: %v", noColor, err)
		return fmt.Errorf("error getting Port status: %w", err)
	}

	if port == nil {
		output.PrintError("No port found with UID: %s", noColor, portUID)
		return fmt.Errorf("no port found with UID: %s", portUID)
	}

	status := []PortStatus{
		{
			UID:    port.UID,
			Name:   port.Name,
			Status: port.ProvisioningStatus,
			Type:   port.Type,
			Speed:  port.PortSpeed,
		},
	}

	return output.PrintOutput(status, outputFormat, noColor)
}

func UpdatePort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return err
	}

	portUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	var req *megaport.ModifyPortRequest

	getSpinner := output.PrintResourceGetting("Port", portUID, noColor)

	originalPort, err := getPortFunc(ctx, client, portUID)

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
		if err == nil {
			req.PortID = portUID
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

	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	updateSpinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := updatePortFunc(ctx, client, req)

	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to update port: %v", noColor, err)
		return err
	}

	if !resp.IsUpdated {
		output.PrintError("Port update request was not successful", noColor)
		return fmt.Errorf("port update request was not successful")
	}

	output.PrintResourceUpdated("Port", portUID, noColor)

	getUpdatedSpinner := output.PrintResourceGetting("Port", portUID, noColor)

	updatedPort, err := getPortFunc(ctx, client, portUID)

	getUpdatedSpinner.Stop()

	if err != nil {
		output.PrintError("Port was updated but failed to retrieve updated details: %v", noColor, err)
		return nil
	}

	displayPortChanges(originalPort, updatedPort, noColor)

	return nil
}

func DeletePort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	portUID := args[0]

	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		output.PrintError("Failed to get delete now flag: %v", noColor, err)
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		output.PrintError("Failed to get force flag: %v", noColor, err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete port " + portUID + "? "
		if !utils.ConfirmPrompt(confirmMsg, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	safeDelete, err := cmd.Flags().GetBool("safe-delete")
	if err != nil {
		output.PrintError("Failed to get safe-delete flag: %v", noColor, err)
		return err
	}

	deleteRequest := &megaport.DeletePortRequest{
		PortID:     portUID,
		DeleteNow:  deleteNow,
		SafeDelete: safeDelete,
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceDeleting("Port", portUID, noColor)

	resp, err := deletePortFunc(ctx, client, deleteRequest)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "Port", portUID)
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := restorePortFunc(ctx, client, portUID)

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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := lockPortFunc(ctx, client, portUID)

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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	portUID := args[0]
	formattedUID := output.FormatUID(portUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceUpdating("Port", portUID, noColor)

	resp, err := unlockPortFunc(ctx, client, portUID)

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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	portUID := args[0]
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		output.PrintError("Invalid VLAN ID: %v", noColor, err)
		return fmt.Errorf("invalid VLAN ID")
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	available, err := checkPortVLANAvailabilityFunc(ctx, client, portUID, vlan)

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

func ListPortResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	portUID := args[0]
	return utils.ListResourceTags("Port", portUID, noColor, outputFormat, func(ctx context.Context, uid string) (map[string]string, error) {
		client, err := config.LoginFunc(ctx)
		if err != nil {
			return nil, err
		}
		return listPortResourceTagsFunc(ctx, client, uid)
	})
}

func UpdatePortResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	portUID := args[0]
	var client *megaport.Client
	login := func(ctx context.Context) error {
		var err error
		client, err = config.LoginFunc(ctx)
		return err
	}
	return utils.UpdateResourceTags(utils.UpdateTagsOptions{
		ResourceType:  "Port",
		UID:           portUID,
		NoColor:       noColor,
		Cmd:           cmd,
		ExtraTagFlags: true,
		ListFunc: func(ctx context.Context, uid string) (map[string]string, error) {
			if err := login(ctx); err != nil {
				return nil, err
			}
			return client.PortService.ListPortResourceTags(ctx, uid)
		},
		UpdateFunc: func(ctx context.Context, uid string, tags map[string]string) error {
			return client.PortService.UpdatePortResourceTags(ctx, uid, tags)
		},
	})
}
