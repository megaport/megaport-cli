package ports

import (
	"context"
	"encoding/json"
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

func exportPortConfig(port *megaport.Port) map[string]interface{} {
	m := map[string]interface{}{
		"name":                  port.Name,
		"term":                  port.ContractTermMonths,
		"portSpeed":             port.PortSpeed,
		"locationId":            port.LocationID,
		"marketPlaceVisibility": port.MarketplaceVisibility,
	}
	if port.DiversityZone != "" {
		m["diversityZone"] = port.DiversityZone
	}
	if port.CostCentre != "" {
		m["costCentre"] = port.CostCentre
	}
	return m
}

func buildPortRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyPortRequest, error) {
	return utils.ResolveInput(utils.InputConfig[*megaport.BuyPortRequest]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      noColor,
		FlagsProvided: func() bool {
			return cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
				cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
				cmd.Flags().Changed("marketplace-visibility")
		},
		FromJSON:   processJSONPortInput,
		FromFlags:  func() (*megaport.BuyPortRequest, error) { return processFlagPortInput(cmd) },
		FromPrompt: func() (*megaport.BuyPortRequest, error) { return promptForPortDetails(noColor) },
	})
}

func buildLAGPortRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyPortRequest, error) {
	return utils.ResolveInput(utils.InputConfig[*megaport.BuyPortRequest]{
		ResourceName: "port",
		Cmd:          cmd,
		NoColor:      noColor,
		FlagsProvided: func() bool {
			return cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
				cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
				cmd.Flags().Changed("lag-count") || cmd.Flags().Changed("marketplace-visibility")
		},
		FromJSON:   processJSONPortInput,
		FromFlags:  func() (*megaport.BuyPortRequest, error) { return processFlagLAGPortInput(cmd) },
		FromPrompt: func() (*megaport.BuyPortRequest, error) { return promptForLAGPortDetails(noColor) },
	})
}

func BuyPort(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
	defer cancel()

	req, err := buildPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = utils.DefaultProvisionTimeout
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

	var resp *megaport.BuyPortResponse
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = buyPortFunc(ctx, client, req)
		return e
	})

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
	defer cancel()

	req, err := buildLAGPortRequest(cmd, noColor)
	if err != nil {
		return err
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = utils.DefaultProvisionTimeout
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
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("Port", noColor)

	ports, err := listPortsFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list ports: %v", noColor, err)
		return fmt.Errorf("failed to list ports: %w", err)
	}

	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	filteredPorts := filterPorts(ports, locationID, portSpeed, portName, includeInactive)

	limit, _ := cmd.Flags().GetInt("limit")

	tagFilters, _ := cmd.Flags().GetStringArray("tag")
	if len(tagFilters) > 0 {
		tagSpinner := output.PrintCustomSpinner("Fetching tags for", "ports", noColor)
		filteredPorts = utils.ApplyTagFilter(ctx, filteredPorts,
			func(p *megaport.Port) string { return p.UID },
			func(ctx context.Context, uid string) (map[string]string, error) {
				return listPortResourceTagsFunc(ctx, client, uid)
			},
			tagFilters, limit,
			func(uid string, err error) {
				output.PrintWarning("Failed to fetch tags for port %s, skipping: %v", noColor, uid, err)
			},
		)
		tagSpinner.Stop()
	}
	return utils.ApplyLimitAndPrint(filteredPorts, limit, outputFormat, noColor,
		"No ports found. Create one with 'megaport ports buy'.", printPorts)
}

func GetPort(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchGetPort(cmd, args, noColor, outputFormat)
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := getPortFunc(ctx, client, portUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "Port", portUID)
		output.PrintError("Failed to get port: %v", noColor, err)
		return fmt.Errorf("failed to get port: %w", err)
	}

	if port == nil {
		output.PrintError("No port found with UID: %s", noColor, portUID)
		return fmt.Errorf("no port found with UID: %s", portUID)
	}

	export, _ := cmd.Flags().GetBool("export")
	if export {
		cfg := exportPortConfig(port)
		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal export config: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	err = printPorts([]*megaport.Port{port}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print ports: %v", noColor, err)
		return fmt.Errorf("failed to print ports: %w", err)
	}
	return nil
}

func watchGetPort(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	portUID := args[0]
	return utils.WatchResource(cmd, "Port", portUID, noColor, outputFormat, config.Login,
		func(pollCtx context.Context, client *megaport.Client) (string, error) {
			port, err := getPortFunc(pollCtx, client, portUID)
			if err != nil {
				return "", err
			}
			if port == nil {
				return "", fmt.Errorf("no port found with UID: %s", portUID)
			}
			err = printPorts([]*megaport.Port{port}, outputFormat, noColor)
			return port.ProvisioningStatus, err
		})
}

func GetPortStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchPortStatus(cmd, args, noColor, outputFormat)
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	portUID := args[0]

	spinner := output.PrintResourceGetting("Port", portUID, noColor)

	port, err := getPortFunc(ctx, client, portUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "Port", portUID)
		output.PrintError("Failed to get Port status: %v", noColor, err)
		return fmt.Errorf("failed to get Port status: %w", err)
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

func watchPortStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	portUID := args[0]
	return utils.WatchResource(cmd, "Port", portUID, noColor, outputFormat, config.Login,
		func(pollCtx context.Context, client *megaport.Client) (string, error) {
			port, err := getPortFunc(pollCtx, client, portUID)
			if err != nil {
				return "", err
			}
			if port == nil {
				return "", fmt.Errorf("no port found with UID: %s", portUID)
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
			err = output.PrintOutput(status, outputFormat, noColor)
			return port.ProvisioningStatus, err
		})
}
