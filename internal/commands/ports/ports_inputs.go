package ports

import (
	"encoding/json"
	"fmt"

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
	var resourceTags map[string]string
	if resourceTagsStr != "" {
		if err := json.Unmarshal([]byte(resourceTagsStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse resource tags JSON: %v", err)
		}
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

func processJSONUpdatePortInput(jsonStr, jsonFile string) (*megaport.ModifyPortRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.ModifyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if req.ContractTermMonths != nil {
		if err := validation.ValidateContractTerm(*req.ContractTermMonths); err != nil {
			return nil, err
		}
	}

	isUpdating := req.Name != "" ||
		req.MarketplaceVisibility != nil ||
		req.CostCentre != "" ||
		req.ContractTermMonths != nil

	if !isUpdating {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}

func processFlagUpdatePortInput(cmd *cobra.Command, portUID string) (*megaport.ModifyPortRequest, error) {
	req := &megaport.ModifyPortRequest{
		PortID: portUID,
	}

	nameSet := cmd.Flags().Changed("name")
	mvSet := cmd.Flags().Changed("marketplace-visibility")
	ccSet := cmd.Flags().Changed("cost-centre")
	termSet := cmd.Flags().Changed("term")

	if !nameSet && !mvSet && !ccSet && !termSet {
		return nil, fmt.Errorf("at least one field must be updated")
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
		if term != 0 {
			if err := validation.ValidateContractTerm(term); err != nil {
				return nil, err
			}
			req.ContractTermMonths = &term
		}
	}

	return req, nil
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
	var resourceTags map[string]string
	if resourceTagsStr != "" {
		if err := json.Unmarshal([]byte(resourceTagsStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse resource tags JSON: %v", err)
		}
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
		return nil, err
	}

	req := &megaport.BuyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if err := validation.ValidatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}
