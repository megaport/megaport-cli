package servicekeys

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func buildCreateServiceKeyRequest(cmd *cobra.Command, noColor bool) (*megaport.CreateServiceKeyRequest, error) {
	return utils.ResolveInput(utils.InputConfig[*megaport.CreateServiceKeyRequest]{
		ResourceName: "service key",
		Cmd:          cmd,
		NoColor:      noColor,
		FlagsProvided: func() bool {
			return cmd.Flags().Changed("product-uid") || cmd.Flags().Changed("product-id") ||
				cmd.Flags().Changed("single-use") || cmd.Flags().Changed("max-speed") ||
				cmd.Flags().Changed("description") || cmd.Flags().Changed("start-date") ||
				cmd.Flags().Changed("end-date") || cmd.Flags().Changed("active") ||
				cmd.Flags().Changed("pre-approved") || cmd.Flags().Changed("vlan")
		},
		FromJSON:   processJSONCreateServiceKeyInput,
		FromFlags:  func() (*megaport.CreateServiceKeyRequest, error) { return processFlagCreateServiceKeyInput(cmd) },
		FromPrompt: func() (*megaport.CreateServiceKeyRequest, error) { return promptForCreateServiceKeyDetails(noColor) },
	})
}

func CreateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	req, err := buildCreateServiceKeyRequest(cmd, noColor)
	if err != nil {
		return err
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceCreating("Service Key", req.Description, noColor)

	resp, err := client.ServiceKeyService.CreateServiceKey(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to create service key: %v", noColor, err)
		return fmt.Errorf("failed to create service key: %w", err)
	}

	if resp == nil {
		output.PrintError("Service key create returned an empty API response", noColor)
		return fmt.Errorf("empty response from API")
	}

	output.PrintResourceCreated("Service Key", resp.ServiceKeyUID, noColor)
	return nil
}

func UpdateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	key := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	if err := utils.CheckInteractiveConflict(interactive, utils.HasConflictingInputFlags(cmd)); err != nil {
		output.PrintError("%v", noColor, err)
		return err
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	// SingleUse and Active are always serialized by the SDK (no omitempty),
	// so every input mode merges from the current key to avoid resetting
	// fields the user didn't ask to change.
	current, err := client.ServiceKeyService.GetServiceKey(ctx, key)
	if err != nil {
		output.PrintError("Failed to fetch current service key: %v", noColor, err)
		return fmt.Errorf("failed to fetch current service key: %w", err)
	}

	if current == nil {
		output.PrintError("Service key get returned an empty API response", noColor)
		return fmt.Errorf("empty response from API")
	}

	var req *megaport.UpdateServiceKeyRequest
	switch {
	case jsonStr != "" || jsonFile != "":
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildUpdateServiceKeyRequestFromJSON(jsonStr, jsonFile, key, current)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	case interactive:
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = promptForUpdateServiceKeyDetails(key, current, noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	default:
		req, err = buildUpdateServiceKeyRequestFromFlags(cmd, key, current)
		if err != nil {
			output.PrintError("Failed to process flags: %v", noColor, err)
			return err
		}
	}

	spinner := output.PrintResourceUpdating("Service Key", key, noColor)

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update service key: %v", noColor, err)
		return fmt.Errorf("failed to update service key: %w", err)
	}

	if resp == nil {
		output.PrintError("Service key update returned an empty API response", noColor)
		return fmt.Errorf("empty response from API")
	}

	if !resp.IsUpdated {
		output.PrintError("Service key update was not applied", noColor)
		return fmt.Errorf("service key update was not applied")
	}

	output.PrintResourceUpdated("Service Key", key, noColor)
	return nil
}

func ListServiceKeys(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("Service Key", noColor)

	req := &megaport.ListServiceKeysRequest{}
	if cmd.Flags().Changed("product-uid") {
		productUID, _ := cmd.Flags().GetString("product-uid")
		req.ProductUID = &productUID
	}
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list service keys: %v", noColor, err)
		return fmt.Errorf("failed to list service keys: %w", err)
	}

	if resp == nil {
		output.PrintError("Service key list returned an empty API response", noColor)
		return fmt.Errorf("empty response from API")
	}

	serviceKeys := resp.ServiceKeys

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(serviceKeys) > limit {
		serviceKeys = serviceKeys[:limit]
	}

	if len(serviceKeys) == 0 && outputFormat == utils.FormatTable {
		output.PrintInfo("No service keys found.", noColor)
		return nil
	}

	outputs := make([]serviceKeyOutput, 0, len(serviceKeys))
	for _, sk := range serviceKeys {
		op, err := toServiceKeyOutput(sk)
		if err != nil {
			output.PrintError("Failed to convert service key: %v", noColor, err)
			return fmt.Errorf("failed to convert service key: %w", err)
		}
		outputs = append(outputs, op)
	}

	return output.PrintOutput(outputs, outputFormat, noColor)
}

func GetServiceKey(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	keyID := args[0]

	spinner := output.PrintResourceGetting("Service Key", keyID, noColor)

	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get service key: %v", noColor, err)
		return fmt.Errorf("failed to get service key: %w", err)
	}

	if resp == nil {
		output.PrintError("Service key get returned an empty API response", noColor)
		return fmt.Errorf("empty response from API")
	}

	op, err := toServiceKeyOutput(resp)
	if err != nil {
		output.PrintError("Failed to convert service key: %v", noColor, err)
		return fmt.Errorf("failed to convert service key: %w", err)
	}
	return output.PrintOutput([]serviceKeyOutput{op}, outputFormat, noColor)
}
