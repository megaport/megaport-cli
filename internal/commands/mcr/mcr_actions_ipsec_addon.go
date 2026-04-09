package mcr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// AddMCRIPSecAddOn adds an IPSec add-on to an existing MCR.
func AddMCRIPSecAddOn(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	mcrUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	flagsProvided := cmd.Flags().Changed("tunnel-count")

	var tunnelCount int
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		tunnelCount, err = parseIPSecTunnelCountFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		if tunnelCount, err = cmd.Flags().GetInt("tunnel-count"); err != nil {
			return fmt.Errorf("invalid tunnel-count flag: %w", err)
		}
	} else if interactive {
		tunnelCount, err = promptForIPSecTunnelCount(noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or --tunnel-count to specify IPSec add-on details")
	}

	// allowZeroDisable=false: for add, 0 means "use API default of 10", not disable
	if tunnelCount != 0 {
		if err := validation.ValidateIPSecTunnelCount(tunnelCount, false); err != nil {
			return err
		}
	}

	req := megaport.MCRAddOnRequest{
		AddOn: &megaport.MCRAddOnIPsecConfig{
			AddOnType:   megaport.AddOnTypeIPsec,
			TunnelCount: tunnelCount,
		},
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceCreating("IPSec Add-On", mcrUID, noColor)
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		return updateMCRWithAddOnFunc(ctx, client, mcrUID, req)
	})
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to add IPSec add-on: %v", noColor, err)
		return err
	}

	output.PrintSuccess("IPSec add-on added successfully to MCR: %s", noColor, mcrUID)
	return nil
}

// UpdateMCRIPSecAddOn updates an existing IPSec add-on on an MCR.
// Setting tunnel-count to 0 disables IPSec.
func UpdateMCRIPSecAddOn(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	mcrUID := args[0]
	addOnUID := args[1]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	flagsProvided := cmd.Flags().Changed("tunnel-count")

	var tunnelCount int
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		tunnelCount, err = parseIPSecTunnelCountFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		if tunnelCount, err = cmd.Flags().GetInt("tunnel-count"); err != nil {
			return fmt.Errorf("invalid tunnel-count flag: %w", err)
		}
	} else if interactive {
		tunnelCount, err = promptForIPSecTunnelCountUpdate(noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or --tunnel-count to specify tunnel count")
	}

	// allowZeroDisable=true: for update, 0 means disable IPSec
	if err := validation.ValidateIPSecTunnelCount(tunnelCount, true); err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceUpdating("IPSec Add-On", addOnUID, noColor)
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		return updateMCRIPsecAddOnFunc(ctx, client, mcrUID, addOnUID, tunnelCount)
	})
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update IPSec add-on: %v", noColor, err)
		return err
	}

	if tunnelCount == 0 {
		output.PrintSuccess("IPSec add-on disabled successfully", noColor)
	} else {
		output.PrintSuccess("IPSec add-on updated successfully - tunnel count: %d", noColor, tunnelCount)
	}
	return nil
}

// parseIPSecTunnelCountFromJSON parses a tunnel count from JSON input.
// Expects {"tunnelCount": N}.
func parseIPSecTunnelCountFromJSON(jsonStr, jsonFile string) (int, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return 0, err
	}
	var data struct {
		TunnelCount int `json:"tunnelCount"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return data.TunnelCount, nil
}
