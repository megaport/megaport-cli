package ports

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

func promptForPortDetails(noColor bool) (*megaport.BuyPortRequest, error) {
	req := &megaport.BuyPortRequest{}

	name, err := utils.ResourcePrompt("port", "Enter port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidatePortName(name); err != nil {
		return nil, err
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter term (%s) (required): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %w", err)
	}
	if err := validation.ValidateContractTerm(term); err != nil {
		return nil, err
	}
	req.Term = term

	portSpeedStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter port speed (%s) (required): ", validation.FormatIntSlice(validation.ValidPortSpeeds)), noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || !slices.Contains(validation.ValidPortSpeeds, portSpeed) {
		return nil, fmt.Errorf("invalid port speed, must be one of %s", validation.FormatIntSlice(validation.ValidPortSpeeds))
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
	if err := validation.ValidatePortName(name); err != nil {
		return nil, err
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter term (%s) (required): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %w", err)
	}
	if err := validation.ValidateContractTerm(term); err != nil {
		return nil, err
	}
	req.Term = term

	portSpeedStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter port speed (%s) (required): ", validation.FormatIntSlice(validation.ValidLAGPortSpeeds)), noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || !slices.Contains(validation.ValidLAGPortSpeeds, portSpeed) {
		return nil, fmt.Errorf("invalid port speed, must be one of %s", validation.FormatIntSlice(validation.ValidLAGPortSpeeds))
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

	lagCountStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter LAG count (%d-%d) (required): ", validation.MinLAGCount, validation.MaxLAGCount), noColor)
	if err != nil {
		return nil, err
	}
	lagCount, err := strconv.Atoi(lagCountStr)
	if err != nil || lagCount < validation.MinLAGCount || lagCount > validation.MaxLAGCount {
		return nil, fmt.Errorf("invalid LAG count, must be between %d and %d", validation.MinLAGCount, validation.MaxLAGCount)
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

	termStr, err := utils.ResourcePrompt("port", fmt.Sprintf("Enter new term (%s) (optional): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, err
	}
	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil {
			return nil, fmt.Errorf("invalid term: %w", err)
		}
		if err := validation.ValidateContractTerm(term); err != nil {
			return nil, err
		}
		req.ContractTermMonths = &term
	}

	if req.Name == "" && req.MarketplaceVisibility == nil && req.CostCentre == "" && req.ContractTermMonths == nil {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}
