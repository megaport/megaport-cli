package vxc

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
	req := &megaport.UpdateVXCRequest{}

	// Handle simple string and int fields
	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		req.Name = &name
	}

	if cmd.Flags().Changed("rate-limit") {
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		if rateLimit < 0 {
			return nil, fmt.Errorf("rate-limit must be greater than or equal to 0")
		}
		req.RateLimit = &rateLimit
	}

	if cmd.Flags().Changed("term") {
		term, _ := cmd.Flags().GetInt("term")
		if term != 0 {
			if err := validation.ValidateContractTerm(term); err != nil {
				return nil, err
			}
		}
		req.Term = &term
	}

	if cmd.Flags().Changed("cost-centre") {
		costCentre, _ := cmd.Flags().GetString("cost-centre")
		req.CostCentre = &costCentre
	}

	if cmd.Flags().Changed("shutdown") {
		shutdown, _ := cmd.Flags().GetBool("shutdown")
		req.Shutdown = &shutdown
	}

	// Handle VLAN fields
	if cmd.Flags().Changed("a-end-vlan") {
		aEndVLAN, _ := cmd.Flags().GetInt("a-end-vlan")
		if err := validation.ValidateVLAN(aEndVLAN); err != nil {
			return nil, fmt.Errorf("a-end-vlan: %w", err)
		}
		req.AEndVLAN = &aEndVLAN
	}

	if cmd.Flags().Changed("b-end-vlan") {
		bEndVLAN, _ := cmd.Flags().GetInt("b-end-vlan")
		if err := validation.ValidateVLAN(bEndVLAN); err != nil {
			return nil, fmt.Errorf("b-end-vlan: %w", err)
		}
		req.BEndVLAN = &bEndVLAN
	}

	if cmd.Flags().Changed("a-end-inner-vlan") {
		aEndInnerVLAN, _ := cmd.Flags().GetInt("a-end-inner-vlan")
		if err := validation.ValidateVXCEndInnerVLAN(aEndInnerVLAN); err != nil {
			return nil, fmt.Errorf("invalid a-end-inner-vlan: %w", err)
		}
		req.AEndInnerVLAN = &aEndInnerVLAN
	}

	if cmd.Flags().Changed("b-end-inner-vlan") {
		bEndInnerVLAN, _ := cmd.Flags().GetInt("b-end-inner-vlan")
		if err := validation.ValidateVXCEndInnerVLAN(bEndInnerVLAN); err != nil {
			return nil, fmt.Errorf("invalid b-end-inner-vlan: %w", err)
		}
		req.BEndInnerVLAN = &bEndInnerVLAN
	}

	// Handle product UIDs
	if cmd.Flags().Changed("a-end-uid") {
		aEndUID, _ := cmd.Flags().GetString("a-end-uid")
		req.AEndProductUID = &aEndUID
	}

	if cmd.Flags().Changed("b-end-uid") {
		bEndUID, _ := cmd.Flags().GetString("b-end-uid")
		req.BEndProductUID = &bEndUID
	}

	// Handle approval and vNIC index fields
	if cmd.Flags().Changed("is-approved") {
		isApproved, _ := cmd.Flags().GetBool("is-approved")
		req.IsApproved = &isApproved
	}
	if cmd.Flags().Changed("a-vnic-index") {
		aVnicIndex, _ := cmd.Flags().GetInt("a-vnic-index")
		if err := validation.ValidateVNICIndex(aVnicIndex); err != nil {
			return nil, fmt.Errorf("invalid a-vnic-index: %w", err)
		}
		req.AVnicIndex = &aVnicIndex
	}
	if cmd.Flags().Changed("b-vnic-index") {
		bVnicIndex, _ := cmd.Flags().GetInt("b-vnic-index")
		if err := validation.ValidateVNICIndex(bVnicIndex); err != nil {
			return nil, fmt.Errorf("invalid b-vnic-index: %w", err)
		}
		req.BVnicIndex = &bVnicIndex
	}

	// Handle partner configurations
	if cmd.Flags().Changed("a-end-partner-config") {
		aEndPartnerConfigStr, _ := cmd.Flags().GetString("a-end-partner-config")
		if aEndPartnerConfigStr != "" {
			aEndPartnerConfig, err := parsePartnerConfigFromJSON(aEndPartnerConfigStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse a-end-partner-config: %w", err)
			}

			// Verify it's a VRouter config which is the only updatable partner config
			if _, ok := aEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig); !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			req.AEndPartnerConfig = aEndPartnerConfig
		}
	}

	if cmd.Flags().Changed("b-end-partner-config") {
		bEndPartnerConfigStr, _ := cmd.Flags().GetString("b-end-partner-config")
		if bEndPartnerConfigStr != "" {
			bEndPartnerConfig, err := parsePartnerConfigFromJSON(bEndPartnerConfigStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse b-end-partner-config: %w", err)
			}

			// Verify it's a VRouter config which is the only updatable partner config
			if _, ok := bEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig); !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			req.BEndPartnerConfig = bEndPartnerConfig
		}
	}

	return req, nil
}

var buildUpdateVXCRequestFromJSON = func(jsonStr string, jsonFilePath string) (*megaport.UpdateVXCRequest, error) {
	if jsonStr == "" && jsonFilePath == "" {
		return nil, fmt.Errorf("either json or json-file must be provided")
	}

	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFilePath)
	if err != nil {
		return nil, err
	}

	// Parse raw JSON first to handle partner configs
	var rawData map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	req := &megaport.UpdateVXCRequest{}

	if rateLimit, ok := rawData["rateLimit"].(float64); ok {
		rateLimitInt := int(rateLimit)
		if rateLimitInt < 0 {
			return nil, fmt.Errorf("rateLimit must be greater than or equal to 0")
		}
		req.RateLimit = &rateLimitInt
	}

	if term, ok := rawData["term"].(float64); ok {
		if term != math.Trunc(term) {
			return nil, fmt.Errorf("term must be a whole number, got %v", term)
		}
		termInt := int(term)
		if termInt != 0 {
			if err := validation.ValidateContractTerm(termInt); err != nil {
				return nil, err
			}
		}
		req.Term = &termInt
	}

	if costCentre, ok := rawData["costCentre"].(string); ok {
		req.CostCentre = &costCentre
	}

	if shutdown, ok := rawData["shutdown"].(bool); ok {
		req.Shutdown = &shutdown
	}

	// Handle nested configurations in addition to flat fields
	if aEndConfig, ok := rawData["aEndConfiguration"].(map[string]interface{}); ok {
		if vlan, ok := aEndConfig["vlan"].(float64); ok {
			vlanInt := int(vlan)
			if err := validation.ValidateVLAN(vlanInt); err != nil {
				return nil, fmt.Errorf("aEndConfiguration.vlan: %w", err)
			}
			req.AEndVLAN = &vlanInt
		}
	} else {
		if aEndVLAN, ok := rawData["aEndVlan"].(float64); ok {
			aEndVLANInt := int(aEndVLAN)
			if err := validation.ValidateVLAN(aEndVLANInt); err != nil {
				return nil, fmt.Errorf("aEndVlan: %w", err)
			}
			req.AEndVLAN = &aEndVLANInt
		}
	}

	if bEndConfig, ok := rawData["bEndConfiguration"].(map[string]interface{}); ok {
		if vlan, ok := bEndConfig["vlan"].(float64); ok {
			vlanInt := int(vlan)
			if err := validation.ValidateVLAN(vlanInt); err != nil {
				return nil, fmt.Errorf("bEndConfiguration.vlan: %w", err)
			}
			req.BEndVLAN = &vlanInt
		}
	} else {
		if bEndVLAN, ok := rawData["bEndVlan"].(float64); ok {
			bEndVLANInt := int(bEndVLAN)
			if err := validation.ValidateVLAN(bEndVLANInt); err != nil {
				return nil, fmt.Errorf("bEndVlan: %w", err)
			}
			req.BEndVLAN = &bEndVLANInt
		}
	}

	// Handle VXC name field variants
	if name, ok := rawData["name"].(string); ok {
		req.Name = &name
	} else if vxcName, ok := rawData["vxcName"].(string); ok {
		req.Name = &vxcName
	}

	if aEndInnerVLAN, ok := rawData["aEndInnerVlan"].(float64); ok {
		aEndInnerVLANInt := int(aEndInnerVLAN)
		if err := validation.ValidateVXCEndInnerVLAN(aEndInnerVLANInt); err != nil {
			return nil, fmt.Errorf("invalid aEndInnerVlan: %w", err)
		}
		req.AEndInnerVLAN = &aEndInnerVLANInt
	}

	if bEndInnerVLAN, ok := rawData["bEndInnerVlan"].(float64); ok {
		bEndInnerVLANInt := int(bEndInnerVLAN)
		if err := validation.ValidateVXCEndInnerVLAN(bEndInnerVLANInt); err != nil {
			return nil, fmt.Errorf("invalid bEndInnerVlan: %w", err)
		}
		req.BEndInnerVLAN = &bEndInnerVLANInt
	}

	// Handle product UIDs
	if aEndUID, ok := rawData["aEndUid"].(string); ok {
		req.AEndProductUID = &aEndUID
	}

	if bEndUID, ok := rawData["bEndUid"].(string); ok {
		req.BEndProductUID = &bEndUID
	}

	// Handle partner configurations - using direct map access
	if aEndPartnerConfigRaw, ok := rawData["aEndPartnerConfig"].(map[string]interface{}); ok {
		if connectType, ok := aEndPartnerConfigRaw["connectType"].(string); ok && strings.ToUpper(connectType) == "VROUTER" {
			aEndPartnerConfig, err := parsePartnerConfigFromMap(aEndPartnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse A-End partner config: %w", err)
			}

			req.AEndPartnerConfig = aEndPartnerConfig
		} else {
			return nil, fmt.Errorf("only VRouter partner configurations can be updated")
		}
	}

	if bEndPartnerConfigRaw, ok := rawData["bEndPartnerConfig"].(map[string]interface{}); ok {
		if connectType, ok := bEndPartnerConfigRaw["connectType"].(string); ok && strings.ToUpper(connectType) == "VROUTER" {
			bEndPartnerConfig, err := parsePartnerConfigFromMap(bEndPartnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse B-End partner config: %w", err)
			}

			req.BEndPartnerConfig = bEndPartnerConfig
		} else {
			return nil, fmt.Errorf("only VRouter partner configurations can be updated")
		}
	}

	// Handle approval and vNIC index fields from JSON
	if isApproved, ok := rawData["isApproved"].(bool); ok {
		req.IsApproved = &isApproved
	}
	if aVnicIndex, ok := rawData["aVnicIndex"].(float64); ok {
		if aVnicIndex != math.Trunc(aVnicIndex) {
			return nil, fmt.Errorf("aVnicIndex must be a whole number, got %v", aVnicIndex)
		}
		idx := int(aVnicIndex)
		if err := validation.ValidateVNICIndex(idx); err != nil {
			return nil, fmt.Errorf("invalid aVnicIndex: %w", err)
		}
		req.AVnicIndex = &idx
	}
	if bVnicIndex, ok := rawData["bVnicIndex"].(float64); ok {
		if bVnicIndex != math.Trunc(bVnicIndex) {
			return nil, fmt.Errorf("bVnicIndex must be a whole number, got %v", bVnicIndex)
		}
		idx := int(bVnicIndex)
		if err := validation.ValidateVNICIndex(idx); err != nil {
			return nil, fmt.Errorf("invalid bVnicIndex: %w", err)
		}
		req.BVnicIndex = &idx
	}

	// Set wait for update to true with a reasonable timeout
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	return req, nil
}
