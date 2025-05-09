package ports

import (
	"fmt"
	"strconv"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func promptForPortDetails(noColor bool) (*megaport.BuyPortRequest, error) {
	req := &megaport.BuyPortRequest{}

	name, err := utils.ResourcePrompt("port", "Enter port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("port name is required")
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("port", "Enter term (1, 12, 24, 36) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	req.Term = term

	portSpeedStr, err := utils.ResourcePrompt("port", "Enter port speed (1000, 10000, 100000) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 1000 && portSpeed != 10000 && portSpeed != 100000) {
		return nil, fmt.Errorf("invalid port speed, must be one of 1000, 10000, 100000")
	}
	req.PortSpeed = portSpeed

	locationIDStr, err := utils.ResourcePrompt("port", "Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID")
	}
	req.LocationId = locationID

	marketplaceVisibilityStr, err := utils.ResourcePrompt("port", "Enter marketplace visibility (true/false) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
	}
	req.MarketPlaceVisibility = marketplaceVisibility

	diversityZone, err := utils.ResourcePrompt("port", "Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.DiversityZone = diversityZone

	costCentre, err := utils.ResourcePrompt("port", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	promoCode, err := utils.ResourcePrompt("port", "Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	resourceTags, err := utils.ResourceTagsPrompt(noColor)
	if err != nil {
		return nil, err
	}
	req.ResourceTags = resourceTags

	return req, nil
}

func promptForLAGPortDetails(noColor bool) (*megaport.BuyPortRequest, error) {
	req := &megaport.BuyPortRequest{}

	name, err := utils.ResourcePrompt("port", "Enter port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("port name is required")
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("port", "Enter term (1, 12, 24, 36) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	req.Term = term

	portSpeedStr, err := utils.ResourcePrompt("port", "Enter port speed (10000 or 100000) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 10000 && portSpeed != 100000) {
		return nil, fmt.Errorf("invalid port speed, must be one of 10000 or 100000")
	}
	req.PortSpeed = portSpeed

	locationIDStr, err := utils.ResourcePrompt("port", "Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID")
	}
	req.LocationId = locationID

	lagCountStr, err := utils.ResourcePrompt("port", "Enter LAG count (1-8) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	lagCount, err := strconv.Atoi(lagCountStr)
	if err != nil || lagCount < 1 || lagCount > 8 {
		return nil, fmt.Errorf("invalid LAG count, must be between 1 and 8")
	}
	req.LagCount = lagCount

	marketplaceVisibilityStr, err := utils.ResourcePrompt("port", "Enter marketplace visibility (true/false) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
	}
	req.MarketPlaceVisibility = marketplaceVisibility

	diversityZone, err := utils.ResourcePrompt("port", "Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.DiversityZone = diversityZone

	costCentre, err := utils.ResourcePrompt("port", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	promoCode, err := utils.ResourcePrompt("port", "Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	resourceTags, err := utils.ResourceTagsPrompt(noColor)
	if err != nil {
		return nil, err
	}
	req.ResourceTags = resourceTags

	return req, nil
}

func promptForUpdatePortDetails(portUID string, noColor bool) (*megaport.ModifyPortRequest, error) {
	req := &megaport.ModifyPortRequest{
		PortID: portUID,
	}

	name, err := utils.ResourcePrompt("port", "Enter new port name (optional, press Enter to keep current name): ", noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = name
	}

	marketplaceVisibilityStr, err := utils.ResourcePrompt("port", "Enter marketplace visibility (true/false) (optional, press Enter to keep current setting): ", noColor)
	if err != nil {
		return nil, err
	}
	if marketplaceVisibilityStr != "" {
		marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
		if err != nil {
			return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
		}
		req.MarketplaceVisibility = &marketplaceVisibility
	}
	costCentre, err := utils.ResourcePrompt("port", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = costCentre
	}

	termStr, err := utils.ResourcePrompt("port", "Enter new term (1, 12, 24, 36) (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
		req.ContractTermMonths = &term
	}

	if req.Name == "" && req.MarketplaceVisibility == nil && req.CostCentre == "" && req.ContractTermMonths == nil {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}
