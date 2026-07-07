package servicekeys

import (
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

func promptForCreateServiceKeyDetails(noColor bool) (*megaport.CreateServiceKeyRequest, error) {
	productUID, err := utils.ResourcePrompt("service key", "Enter product UID (leave empty to use product ID instead): ", noColor)
	if err != nil {
		return nil, err
	}

	var productID int
	if productUID == "" {
		productIDStr, err := utils.ResourcePrompt("service key", "Enter product ID (leave empty to skip): ", noColor)
		if err != nil {
			return nil, err
		}
		if productIDStr != "" {
			productID, err = validation.ParseInt("product ID", productIDStr)
			if err != nil {
				return nil, err
			}
		}
	}

	singleUse := utils.ConfirmPrompt("Make this a single-use service key?", noColor)

	maxSpeedStr, err := utils.ResourcePrompt("service key", "Enter max speed in Mbps (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	var maxSpeed int
	if maxSpeedStr != "" {
		maxSpeed, err = validation.ParseInt("max speed", maxSpeedStr)
		if err != nil {
			return nil, err
		}
	}

	description, err := utils.ResourcePrompt("service key", "Enter description (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}

	startDate, err := utils.ResourcePrompt("service key", "Enter start date, YYYY-MM-DD (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}

	endDate, err := utils.ResourcePrompt("service key", "Enter end date, YYYY-MM-DD (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}

	validFor, err := parseServiceKeyValidFor(startDate, endDate)
	if err != nil {
		return nil, err
	}

	active := utils.ConfirmPrompt("Make the service key available immediately?", noColor)
	preApproved := utils.ConfirmPrompt("Pre-approve the service key for use?", noColor)

	vlanStr, err := utils.ResourcePrompt("service key", "Enter VLAN ID, required for single-use keys (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	var vlan int
	if vlanStr != "" {
		vlan, err = validation.ParseInt("VLAN ID", vlanStr)
		if err != nil {
			return nil, err
		}
	}

	return &megaport.CreateServiceKeyRequest{
		ProductUID:  productUID,
		ProductID:   productID,
		SingleUse:   singleUse,
		MaxSpeed:    maxSpeed,
		Description: description,
		ValidFor:    validFor,
		Active:      active,
		PreApproved: preApproved,
		VLAN:        vlan,
	}, nil
}

// promptForUpdateServiceKeyDetails mirrors the flag/JSON update paths: fields
// the user skips default to the current key's values rather than the zero
// value, preserving the ESD-1272/ESD-1417 merge behavior.
func promptForUpdateServiceKeyDetails(key string, current *megaport.ServiceKey, noColor bool) (*megaport.UpdateServiceKeyRequest, error) {
	req := &megaport.UpdateServiceKeyRequest{
		Key:        key,
		SingleUse:  current.SingleUse,
		Active:     current.Active,
		ProductUID: current.ProductUID,
	}

	productUID, err := utils.ResourcePrompt("service key", "Enter new product UID (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}

	var productIDStr string
	if productUID == "" {
		productIDStr, err = utils.ResourcePrompt("service key", "Enter new product ID (leave empty to skip): ", noColor)
		if err != nil {
			return nil, err
		}
	}

	switch {
	case productUID != "":
		req.ProductUID = productUID
	case productIDStr != "":
		productID, err := validation.ParseInt("product ID", productIDStr)
		if err != nil {
			return nil, err
		}
		req.ProductUID = ""
		req.ProductID = productID
	}

	singleUseAns, err := utils.ResourcePrompt("service key", "Update single-use setting? (yes/no, leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(singleUseAns) == "yes" {
		req.SingleUse = utils.ConfirmPrompt("Enable single-use?", noColor)
	}

	activeAns, err := utils.ResourcePrompt("service key", "Update active setting? (yes/no, leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(activeAns) == "yes" {
		req.Active = utils.ConfirmPrompt("Make the service key active?", noColor)
	}

	return req, nil
}
