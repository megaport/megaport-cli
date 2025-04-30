package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("marketplace-visibility")

	var req *megaport.BuyPortRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
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
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

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

	spinner := output.PrintResourceCreating("Port", req.Name, noColor)

	resp, err := buyPortFunc(ctx, client, req)

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

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("lag-count") || cmd.Flags().Changed("marketplace-visibility")

	var req *megaport.BuyPortRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPortInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagLAGPortInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
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
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

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

	spinner := output.PrintResourceCreating("LAG Port", req.Name, noColor)

	resp, err := buyPortFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy LAG port: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("LAG Port", resp.TechnicalServiceUIDs[0], noColor)
	return nil
}

func ListPorts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("Port", noColor)

	ports, err := client.PortService.ListPorts(ctx)

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
		output.PrintWarning("No ports found matching the specified filters", noColor)
	}

	err = printPorts(filteredPorts, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func GetPort(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := client.PortService.GetPort(ctx, portUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get port: %v", noColor, err)
		return fmt.Errorf("error getting port: %v", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := client.PortService.GetPort(ctx, portUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get Port status: %v", noColor, err)
		return fmt.Errorf("error getting Port status: %v", err)
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
			return nil
		}
	}

	deleteRequest := &megaport.DeletePortRequest{
		PortID:    portUID,
		DeleteNow: deleteNow,
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		return err
	}

	tagsMap, err := client.PortService.ListPortResourceTags(ctx, portUID)

	if err != nil {
		output.PrintError("Error getting resource tags for port %s: %v", noColor, portUID, err)
		return fmt.Errorf("error getting resource tags for port %s: %v", portUID, err)
	}

	tags := make([]output.ResourceTag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tags = append(tags, output.ResourceTag{Key: k, Value: v})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	return output.PrintOutput(tags, outputFormat, noColor)
}

func UpdatePortResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	portUID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	existingTags, err := client.PortService.ListPortResourceTags(ctx, portUID)

	if err != nil {
		output.PrintError("Failed to get existing resource tags: %v", noColor, err)
		return err
	}

	interactive, _ := cmd.Flags().GetBool("interactive")

	var resourceTags map[string]string

	if interactive {
		resourceTags, err = utils.UpdateResourceTagsPrompt(existingTags, noColor)
		if err != nil {
			output.PrintError("Failed to update resource tags", noColor, err)
			return err
		}
	} else {
		jsonStr, _ := cmd.Flags().GetString("json")
		jsonFile, _ := cmd.Flags().GetString("json-file")

		tagsStr, _ := cmd.Flags().GetString("tags")
		tagsFile, _ := cmd.Flags().GetString("tags-file")
		resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")

		if jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON: %v", noColor, err)
				return fmt.Errorf("error parsing JSON: %v", err)
			}
		} else if jsonFile != "" {
			jsonData, err := os.ReadFile(jsonFile)
			if err != nil {
				output.PrintError("Failed to read JSON file: %v", noColor, err)
				return fmt.Errorf("error reading JSON file: %v", err)
			}

			if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON file: %v", noColor, err)
				return fmt.Errorf("error parsing JSON file: %v", err)
			}
		} else if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &resourceTags); err != nil {
				output.PrintError("Failed to parse tags JSON: %v", noColor, err)
				return fmt.Errorf("error parsing tags JSON: %v", err)
			}
		} else if resourceTagsStr != "" {
			if err := json.Unmarshal([]byte(resourceTagsStr), &resourceTags); err != nil {
				output.PrintError("Failed to parse resource-tags JSON: %v", noColor, err)
				return fmt.Errorf("error parsing resource-tags JSON: %v", err)
			}
		} else if tagsFile != "" {
			tagData, err := os.ReadFile(tagsFile)
			if err != nil {
				output.PrintError("Failed to read tags file: %v", noColor, err)
				return fmt.Errorf("error reading tags file: %v", err)
			}
			if err := json.Unmarshal(tagData, &resourceTags); err != nil {
				output.PrintError("Failed to parse tags file JSON: %v", noColor, err)
				return fmt.Errorf("error parsing tags file JSON: %v", err)
			}
		} else {
			output.PrintError("No input provided for tags", noColor)
			return fmt.Errorf("no input provided, use --interactive, --json, --json-file, --tags, --resource-tags, or --tags-file to specify resource tags")
		}
	}

	if len(resourceTags) == 0 {
		fmt.Println("No tags provided. The port will have all existing tags removed.")
	}

	spinner := output.PrintResourceUpdating("Port-Resource-Tags", portUID, noColor)

	err = client.PortService.UpdatePortResourceTags(ctx, portUID, resourceTags)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", noColor, err)
		return fmt.Errorf("failed to update resource tags: %v", err)
	}

	fmt.Printf("Resource tags updated for port %s\n", portUID)
	return nil
}
