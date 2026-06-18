package ix

import (
	"context"
	"fmt"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

func buildIXRequestFromPrompt(_ context.Context, noColor bool) (*megaport.BuyIXRequest, error) {
	productUID, err := utils.ResourcePrompt("ix", "Enter port UID to attach the IX to (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if productUID == "" {
		return nil, fmt.Errorf("product UID is required")
	}

	name, err := utils.ResourcePrompt("ix", "Enter IX name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	networkServiceType, err := utils.ResourcePrompt("ix", "Enter network service type (required, e.g. \"Los Angeles IX\"): ", noColor)
	if err != nil {
		return nil, err
	}
	if networkServiceType == "" {
		return nil, fmt.Errorf("network service type is required")
	}

	asnStr, err := utils.ResourcePrompt("ix", "Enter ASN (required): ", noColor)
	if err != nil {
		return nil, err
	}
	asn, err := validation.ParseInt("ASN", asnStr)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateASN(asn); err != nil {
		return nil, err
	}

	macAddress, err := utils.ResourcePrompt("ix", "Enter MAC address (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateMACAddress(macAddress); err != nil {
		return nil, err
	}

	rateLimitStr, err := utils.ResourcePrompt("ix", "Enter rate limit in Mbps (required): ", noColor)
	if err != nil {
		return nil, err
	}
	rateLimit, err := validation.ParseInt("rate limit", rateLimitStr)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateRateLimit(rateLimit); err != nil {
		return nil, err
	}

	vlanStr, err := utils.ResourcePrompt("ix", "Enter VLAN ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	vlan, err := validation.ParseInt("VLAN", vlanStr)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateVLAN(vlan); err != nil {
		return nil, err
	}

	promoCode, err := utils.ResourcePrompt("ix", "Enter promo code (optional, leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}

	req := &megaport.BuyIXRequest{
		ProductUID:         productUID,
		Name:               name,
		NetworkServiceType: networkServiceType,
		ASN:                asn,
		MACAddress:         macAddress,
		RateLimit:          rateLimit,
		VLAN:               vlan,
		PromoCode:          promoCode,
	}

	return req, nil
}

func buildUpdateIXRequestFromPrompt(_ string, noColor bool) (*megaport.UpdateIXRequest, error) { //nolint:unparam
	req := &megaport.UpdateIXRequest{}
	fieldsUpdated := false

	name, err := utils.ResourcePrompt("ix", "Enter new IX name (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = &name
		fieldsUpdated = true
	}

	rateLimitStr, err := utils.ResourcePrompt("ix", "Enter new rate limit in Mbps (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if rateLimitStr != "" {
		rateLimit, err := validation.ParseInt("rate limit", rateLimitStr)
		if err != nil {
			return nil, err
		}
		if err := validation.ValidateRateLimit(rateLimit); err != nil {
			return nil, err
		}
		req.RateLimit = &rateLimit
		fieldsUpdated = true
	}

	costCentre, err := utils.ResourcePrompt("ix", "Enter new cost centre (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = &costCentre
		fieldsUpdated = true
	}

	vlanStr, err := utils.ResourcePrompt("ix", "Enter new VLAN ID (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if vlanStr != "" {
		vlan, err := validation.ParseInt("VLAN", vlanStr)
		if err != nil {
			return nil, err
		}
		if err := validation.ValidateVLAN(vlan); err != nil {
			return nil, err
		}
		req.VLAN = &vlan
		fieldsUpdated = true
	}

	macAddress, err := utils.ResourcePrompt("ix", "Enter new MAC address (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if macAddress != "" {
		if err := validation.ValidateMACAddress(macAddress); err != nil {
			return nil, err
		}
		req.MACAddress = &macAddress
		fieldsUpdated = true
	}

	asnStr, err := utils.ResourcePrompt("ix", "Enter new ASN (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if asnStr != "" {
		asn, err := validation.ParseInt("ASN", asnStr)
		if err != nil {
			return nil, err
		}
		if err := validation.ValidateASN(asn); err != nil {
			return nil, err
		}
		req.ASN = &asn
		fieldsUpdated = true
	}

	password, err := utils.PasswordPrompt("Enter new BGP password (leave empty to skip):", noColor)
	if err != nil {
		return nil, err
	}
	if password != "" {
		req.Password = &password
		fieldsUpdated = true
	}

	reverseDns, err := utils.ResourcePrompt("ix", "Enter new reverse DNS (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if reverseDns != "" {
		req.ReverseDns = &reverseDns
		fieldsUpdated = true
	}

	if !fieldsUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}
