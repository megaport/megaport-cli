package ports

import (
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func processFlagLAGPortInput(cmd *cobra.Command) (*megaport.BuyPortRequest, error) {
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	lagCount, _ := cmd.Flags().GetInt("lag-count")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")
	resourceTagsFile, _ := cmd.Flags().GetString("resource-tags-file")
	resourceTags, err := utils.ParseResourceTagsFlagOrFile(resourceTagsStr, resourceTagsFile)
	if err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		LagCount:              lagCount,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
		ResourceTags:          resourceTags,
	}

	if err := validation.ValidateLAGPortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Builds and validates the request without any network round-trip so bad input
// fails fast. The bool reports whether the caller supplied costCentre: the SDK
// sends it without omitempty, so the caller re-sends the current value when this
// is false to avoid wiping it (an explicit empty value still clears it).
func processJSONUpdatePortInput(jsonStr, jsonFile string) (*megaport.ModifyPortRequest, bool, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, false, err
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, false, fmt.Errorf("failed to parse JSON: %w", err)
	}

	req := &megaport.ModifyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, false, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if req.ContractTermMonths != nil {
		if err := validation.ValidateContractTerm(*req.ContractTermMonths); err != nil {
			return nil, false, err
		}
	}

	_, costCentreProvided := jsonMap["costCentre"]

	isUpdating := req.Name != "" ||
		req.MarketplaceVisibility != nil ||
		costCentreProvided ||
		req.ContractTermMonths != nil

	if !isUpdating {
		return nil, false, fmt.Errorf("at least one field must be updated")
	}

	return req, costCentreProvided, nil
}

func processFlagUpdatePortInput(cmd *cobra.Command, portUID string) (*megaport.ModifyPortRequest, bool, error) {
	req := &megaport.ModifyPortRequest{
		PortID: portUID,
	}

	nameSet := cmd.Flags().Changed("name")
	mvSet := cmd.Flags().Changed("marketplace-visibility")
	ccSet := cmd.Flags().Changed("cost-centre")
	termSet := cmd.Flags().Changed("term")

	if !nameSet && !mvSet && !ccSet && !termSet {
		return nil, false, fmt.Errorf("at least one field must be updated")
	}

	if nameSet {
		name, _ := cmd.Flags().GetString("name")
		req.Name = name
	}

	if mvSet {
		marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")
		req.MarketplaceVisibility = &marketplaceVisibility
	}

	if ccSet {
		costCentre, _ := cmd.Flags().GetString("cost-centre")
		req.CostCentre = costCentre
	}

	if termSet {
		term, _ := cmd.Flags().GetInt("term")
		if err := validation.ValidateContractTerm(term); err != nil {
			return nil, false, err
		}
		req.ContractTermMonths = &term
	}

	return req, ccSet, nil
}

func processFlagPortInput(cmd *cobra.Command) (*megaport.BuyPortRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")
	resourceTagsFile, _ := cmd.Flags().GetString("resource-tags-file")
	resourceTags, err := utils.ParseResourceTagsFlagOrFile(resourceTagsStr, resourceTagsFile)
	if err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
		ResourceTags:          resourceTags,
	}

	if err := validation.ValidatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processJSONPortInput(jsonStr, jsonFile string) (*megaport.BuyPortRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	req := &megaport.BuyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("failed to parse JSON: %w", err))
	}

	if err := utils.RejectEmptyTagKeys(req.ResourceTags); err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	if err := validation.ValidatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processJSONLAGPortInput(jsonStr, jsonFile string) (*megaport.BuyPortRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	req := &megaport.BuyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("failed to parse JSON: %w", err))
	}

	if err := utils.RejectEmptyTagKeys(req.ResourceTags); err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	// A missing lagCount unmarshals to 0, which ValidateLAGPortRequest rejects,
	// so an order for the wrong product can't slip through the JSON path.
	if err := validation.ValidateLAGPortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}
