package vxc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var buildVXCRequestFromFlags = func(cmd *cobra.Command, ctx context.Context, svc megaport.VXCService) (*megaport.BuyVXCRequest, error) {
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	aEndUID, _ := cmd.Flags().GetString("a-end-uid")

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	rateLimit, _ := cmd.Flags().GetInt("rate-limit")
	if err := validation.ValidateRateLimit(rateLimit); err != nil {
		return nil, err
	}

	term, _ := cmd.Flags().GetInt("term")
	if err := validation.ValidateContractTerm(term); err != nil {
		return nil, err
	}

	// Get optional fields
	aEndVLAN, _ := cmd.Flags().GetInt("a-end-vlan")
	bEndVLAN, _ := cmd.Flags().GetInt("b-end-vlan")
	aEndInnerVLAN, _ := cmd.Flags().GetInt("a-end-inner-vlan")
	bEndInnerVLAN, _ := cmd.Flags().GetInt("b-end-inner-vlan")
	aEndVNICIndex, _ := cmd.Flags().GetInt("a-end-vnic-index")
	bEndVNICIndex, _ := cmd.Flags().GetInt("b-end-vnic-index")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	serviceKey, _ := cmd.Flags().GetString("service-key")
	costCentre, _ := cmd.Flags().GetString("cost-centre")

	// Create the base request
	req := &megaport.BuyVXCRequest{
		VXCName:    name,
		RateLimit:  rateLimit,
		Term:       term,
		PromoCode:  promoCode,
		ServiceKey: serviceKey,
		CostCentre: costCentre,
	}

	// A-End configuration
	aEndConfig := megaport.VXCOrderEndpointConfiguration{
		VLAN: aEndVLAN,
	}

	// Set MVE config if needed
	if aEndInnerVLAN != 0 || aEndVNICIndex > 0 {
		aEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             aEndInnerVLAN,
			NetworkInterfaceIndex: aEndVNICIndex,
		}
	}

	// Parse A-End partner config if provided
	aEndPartnerConfigStr, _ := cmd.Flags().GetString("a-end-partner-config")
	if aEndPartnerConfigStr != "" {
		aEndPartnerConfig, err := parsePartnerConfigFromJSON(aEndPartnerConfigStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse a-end-partner-config: %w", err)
		}
		// If the A End UID is not provided, attempt to look it up from the partner port key
		if aEndUID == "" {
			switch aEndPartnerConfig := aEndPartnerConfig.(type) {
			case *megaport.VXCPartnerConfigAzure:
				if aEndPartnerConfig.ServiceKey == "" {
					return nil, fmt.Errorf("serviceKey is required for Azure configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.ServiceKey, "AZURE")
				if err != nil {
					return nil, fmt.Errorf("failed to look up Azure Partner Port: %w", err)
				}
				aEndUID = uid
			case *megaport.VXCPartnerConfigGoogle:
				if aEndPartnerConfig.PairingKey == "" {
					return nil, fmt.Errorf("pairingKey is required for Google configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.PairingKey, "GOOGLE")
				if err != nil {
					return nil, fmt.Errorf("failed to look up Google Partner Port: %w", err)
				}
				aEndUID = uid
			case *megaport.VXCPartnerConfigOracle:
				if aEndPartnerConfig.VirtualCircuitId == "" {
					return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.VirtualCircuitId, "ORACLE")
				if err != nil {
					return nil, fmt.Errorf("failed to look up Oracle Partner Port: %w", err)
				}
				aEndUID = uid
				aEndConfig.PartnerConfig = aEndPartnerConfig
			}
		}
	}

	req.AEndConfiguration = aEndConfig

	if aEndUID == "" {
		return nil, fmt.Errorf("a-end-uid was neither specified nor could be looked up")
	}

	req.PortUID = aEndUID

	// B-End configuration
	bEndConfig := megaport.VXCOrderEndpointConfiguration{}

	// Parse B-End partner config if provided
	bEndPartnerConfigStr, _ := cmd.Flags().GetString("b-end-partner-config")
	if bEndPartnerConfigStr != "" {
		bEndPartnerConfig, err := parsePartnerConfigFromJSON(bEndPartnerConfigStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse b-end-partner-config: %w", err)
		}
		bEndConfig.PartnerConfig = bEndPartnerConfig
	}

	bEndUID, _ := cmd.Flags().GetString("b-end-uid")

	// Attempt to look up partner port UID if not provided
	if bEndUID == "" {
		switch bEndPartnerConfig := bEndConfig.PartnerConfig.(type) {
		case *megaport.VXCPartnerConfigAzure:
			if bEndPartnerConfig.ServiceKey == "" {
				return nil, fmt.Errorf("serviceKey is required for Azure configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.ServiceKey, "AZURE")
			if err != nil {
				return nil, fmt.Errorf("failed to look up Azure Partner Port: %w", err)
			}
			bEndUID = uid
		case *megaport.VXCPartnerConfigGoogle:
			if bEndPartnerConfig.PairingKey == "" {
				return nil, fmt.Errorf("pairingKey is required for Google configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.PairingKey, "GOOGLE")
			if err != nil {
				return nil, fmt.Errorf("failed to look up Google Partner Port: %w", err)
			}
			bEndUID = uid
		case *megaport.VXCPartnerConfigOracle:
			if bEndPartnerConfig.VirtualCircuitId == "" {
				return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.VirtualCircuitId, "ORACLE")
			if err != nil {
				return nil, fmt.Errorf("failed to look up Oracle Partner Port: %w", err)
			}
			bEndUID = uid
		}
	}

	if bEndUID == "" {
		return nil, fmt.Errorf("b-end-uid was neither provided nor could be looked up")
	}

	bEndConfig.ProductUID = bEndUID
	bEndConfig.VLAN = bEndVLAN

	// Set MVE config if needed
	if bEndInnerVLAN != 0 || bEndVNICIndex > 0 {
		bEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             bEndInnerVLAN,
			NetworkInterfaceIndex: bEndVNICIndex,
		}
	}

	req.BEndConfiguration = bEndConfig

	if err := validation.ValidateVXCRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// parseVXCEndpointConfig parses common endpoint configuration fields (productUID, vlan,
// diversityZone, partnerConfig, MVE config) from a raw JSON map into a VXCOrderEndpointConfiguration.
// The endLabel parameter (e.g. "A-End", "B-End") is used for error messages.
func parseVXCEndpointConfig(endConfigRaw map[string]interface{}, endLabel string) (megaport.VXCOrderEndpointConfiguration, error) {
	config := megaport.VXCOrderEndpointConfiguration{}

	if productUID, ok := endConfigRaw["productUID"].(string); ok {
		config.ProductUID = productUID
	}

	if vlan, ok := endConfigRaw["vlan"].(float64); ok {
		config.VLAN = int(vlan)
	}

	if diversityZone, ok := endConfigRaw["diversityZone"].(string); ok {
		config.DiversityZone = diversityZone
	}

	// Handle partner config - directly use map data
	if partnerConfigRaw, ok := endConfigRaw["partnerConfig"].(map[string]interface{}); ok {
		partnerConfig, err := parsePartnerConfigFromMap(partnerConfigRaw)
		if err != nil {
			return config, fmt.Errorf("failed to parse %s partner config: %w", endLabel, err)
		}

		config.PartnerConfig = partnerConfig
	}

	// Handle MVE config
	innerVLAN, hasInnerVLAN := endConfigRaw["innerVlan"].(float64)
	vNicIndex, hasVNicIndex := endConfigRaw["vNicIndex"].(float64)

	if hasInnerVLAN || hasVNicIndex {
		mveConfig := &megaport.VXCOrderMVEConfig{}

		if hasInnerVLAN {
			mveConfig.InnerVLAN = int(innerVLAN)
		}

		if hasVNicIndex {
			mveConfig.NetworkInterfaceIndex = int(vNicIndex)
		}

		config.VXCOrderMVEConfig = mveConfig
	}

	return config, nil
}

func buildVXCRequestFromJSON(jsonStr string, jsonFilePath string) (*megaport.BuyVXCRequest, error) {
	if jsonStr == "" && jsonFilePath == "" {
		return nil, fmt.Errorf("either json or json-file must be provided")
	}

	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFilePath)
	if err != nil {
		return nil, err
	}

	// Parse raw JSON first to handle partner configs correctly
	var rawData map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	portUID, ok := rawData["portUid"].(string)
	if !ok {
		return nil, validation.NewValidationError("portUid", "", "Port UID is required")
	}

	// Create the base request
	req := &megaport.BuyVXCRequest{
		PortUID: portUID,
	}

	// Set simple fields
	if vxcName, ok := rawData["vxcName"].(string); ok {
		req.VXCName = vxcName
	}

	if rateLimit, ok := rawData["rateLimit"].(float64); ok {
		req.RateLimit = int(rateLimit)
	}

	if term, ok := rawData["term"].(float64); ok {
		req.Term = int(term)
	}

	if shutdown, ok := rawData["shutdown"].(bool); ok {
		req.Shutdown = shutdown
	}

	if promoCode, ok := rawData["promoCode"].(string); ok {
		req.PromoCode = promoCode
	}

	if serviceKey, ok := rawData["serviceKey"].(string); ok {
		req.ServiceKey = serviceKey
	}

	if costCentre, ok := rawData["costCentre"].(string); ok {
		req.CostCentre = costCentre
	}

	// Handle resource tags if they exist
	if resourceTags, ok := rawData["resourceTags"].(map[string]interface{}); ok {
		req.ResourceTags = make(map[string]string)
		for k, v := range resourceTags {
			if strValue, ok := v.(string); ok {
				req.ResourceTags[k] = strValue
			}
		}
	}

	// Handle A-End configuration
	if aEndConfigRaw, ok := rawData["aEndConfiguration"].(map[string]interface{}); ok {
		aEndConfig, err := parseVXCEndpointConfig(aEndConfigRaw, "A-End")
		if err != nil {
			return nil, err
		}
		req.AEndConfiguration = aEndConfig
	}

	// Handle B-End configuration
	if bEndConfigRaw, ok := rawData["bEndConfiguration"].(map[string]interface{}); ok {
		bEndConfig, err := parseVXCEndpointConfig(bEndConfigRaw, "B-End")
		if err != nil {
			return nil, err
		}
		req.BEndConfiguration = bEndConfig
	}

	if err := validation.ValidateVXCRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}
