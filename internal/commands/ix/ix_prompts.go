package ix

import (
	"context"
	"fmt"
	"strconv"

	"github.com/megaport/megaport-cli/internal/utils"
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
	asn, err := strconv.Atoi(asnStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ASN: %v", err)
	}

	macAddress, err := utils.ResourcePrompt("ix", "Enter MAC address (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if macAddress == "" {
		return nil, fmt.Errorf("MAC address is required")
	}

	rateLimitStr, err := utils.ResourcePrompt("ix", "Enter rate limit in Mbps (required): ", noColor)
	if err != nil {
		return nil, err
	}
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit: %v", err)
	}

	vlanStr, err := utils.ResourcePrompt("ix", "Enter VLAN ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	vlan, err := strconv.Atoi(vlanStr)
	if err != nil {
		return nil, fmt.Errorf("invalid VLAN: %v", err)
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
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid rate limit: %v", err)
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
		vlan, err := strconv.Atoi(vlanStr)
		if err != nil {
			return nil, fmt.Errorf("invalid VLAN: %v", err)
		}
		req.VLAN = &vlan
		fieldsUpdated = true
	}

	macAddress, err := utils.ResourcePrompt("ix", "Enter new MAC address (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if macAddress != "" {
		req.MACAddress = &macAddress
		fieldsUpdated = true
	}

	asnStr, err := utils.ResourcePrompt("ix", "Enter new ASN (leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	if asnStr != "" {
		asn, err := strconv.Atoi(asnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ASN: %v", err)
		}
		req.ASN = &asn
		fieldsUpdated = true
	}

	password, err := utils.ResourcePrompt("ix", "Enter new BGP password (leave empty to skip): ", noColor)
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
