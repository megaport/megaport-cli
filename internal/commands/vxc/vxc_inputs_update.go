package vxc

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
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
			vrouterConfig, ok := aEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig)
			if !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			if err := validation.ValidateVrouterPartnerConfig(vrouterConfig); err != nil {
				return nil, err
			}
			req.AEndPartnerConfig = vrouterConfig
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
			vrouterConfig, ok := bEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig)
			if !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			if err := validation.ValidateVrouterPartnerConfig(vrouterConfig); err != nil {
				return nil, err
			}
			req.BEndPartnerConfig = vrouterConfig
		}
	}

	return req, nil
}

var buildUpdateVXCRequestFromJSON = func(jsonStr string, jsonFilePath string) (*megaport.UpdateVXCRequest, error) {
	if jsonStr == "" && jsonFilePath == "" {
		return nil, exitcodes.NewUsageError(fmt.Errorf("either json or json-file must be provided"))
	}

	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFilePath)
	if err != nil {
		return nil, exitcodes.NewUsageError(err)
	}

	// Parse raw JSON first to handle partner configs
	var rawData map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("failed to parse JSON: %w", err))
	}

	req := &megaport.UpdateVXCRequest{}
	fieldSet := false

	if rateLimit, present, err := utils.JSONNumber(rawData, "rateLimit"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("rateLimit: %w", err))
	} else if present {
		if rateLimit != math.Trunc(rateLimit) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("rateLimit must be a whole number, got %v", rateLimit))
		}
		rateLimitInt := int(rateLimit)
		if rateLimitInt < 0 {
			return nil, exitcodes.NewUsageError(fmt.Errorf("rateLimit must be greater than or equal to 0"))
		}
		req.RateLimit = &rateLimitInt
		fieldSet = true
	}

	if term, present, err := utils.JSONNumber(rawData, "term"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("term: %w", err))
	} else if present {
		if term != math.Trunc(term) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("term must be a whole number, got %v", term))
		}
		termInt := int(term)
		if termInt != 0 {
			if err := validation.ValidateContractTerm(termInt); err != nil {
				return nil, exitcodes.NewUsageError(err)
			}
		}
		req.Term = &termInt
		fieldSet = true
	}

	if costCentre, present, err := utils.JSONString(rawData, "costCentre"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("costCentre: %w", err))
	} else if present {
		req.CostCentre = &costCentre
		fieldSet = true
	}

	if shutdown, present, err := utils.JSONBool(rawData, "shutdown"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("shutdown: %w", err))
	} else if present {
		req.Shutdown = &shutdown
		fieldSet = true
	}

	// Handle nested configurations in addition to flat fields
	if aEndConfig, present, err := utils.JSONObject(rawData, "aEndConfiguration"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aEndConfiguration: %w", err))
	} else if present {
		if vlan, vlanPresent, err := utils.JSONNumber(aEndConfig, "vlan"); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndConfiguration.vlan: %w", err))
		} else if vlanPresent {
			if vlan != math.Trunc(vlan) {
				return nil, exitcodes.NewUsageError(fmt.Errorf("aEndConfiguration.vlan must be a whole number, got %v", vlan))
			}
			vlanInt := int(vlan)
			if err := validation.ValidateVLAN(vlanInt); err != nil {
				return nil, exitcodes.NewUsageError(fmt.Errorf("aEndConfiguration.vlan: %w", err))
			}
			req.AEndVLAN = &vlanInt
			fieldSet = true
		}
	} else if aEndVLAN, present, err := utils.JSONNumber(rawData, "aEndVlan"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aEndVlan: %w", err))
	} else if present {
		if aEndVLAN != math.Trunc(aEndVLAN) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndVlan must be a whole number, got %v", aEndVLAN))
		}
		aEndVLANInt := int(aEndVLAN)
		if err := validation.ValidateVLAN(aEndVLANInt); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndVlan: %w", err))
		}
		req.AEndVLAN = &aEndVLANInt
		fieldSet = true
	}

	if bEndConfig, present, err := utils.JSONObject(rawData, "bEndConfiguration"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bEndConfiguration: %w", err))
	} else if present {
		if vlan, vlanPresent, err := utils.JSONNumber(bEndConfig, "vlan"); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndConfiguration.vlan: %w", err))
		} else if vlanPresent {
			if vlan != math.Trunc(vlan) {
				return nil, exitcodes.NewUsageError(fmt.Errorf("bEndConfiguration.vlan must be a whole number, got %v", vlan))
			}
			vlanInt := int(vlan)
			if err := validation.ValidateVLAN(vlanInt); err != nil {
				return nil, exitcodes.NewUsageError(fmt.Errorf("bEndConfiguration.vlan: %w", err))
			}
			req.BEndVLAN = &vlanInt
			fieldSet = true
		}
	} else if bEndVLAN, present, err := utils.JSONNumber(rawData, "bEndVlan"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bEndVlan: %w", err))
	} else if present {
		if bEndVLAN != math.Trunc(bEndVLAN) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndVlan must be a whole number, got %v", bEndVLAN))
		}
		bEndVLANInt := int(bEndVLAN)
		if err := validation.ValidateVLAN(bEndVLANInt); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndVlan: %w", err))
		}
		req.BEndVLAN = &bEndVLANInt
		fieldSet = true
	}

	// Handle VXC name field variants
	if name, present, err := utils.JSONString(rawData, "name"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("name: %w", err))
	} else if present {
		req.Name = &name
		fieldSet = true
	} else if vxcName, present, err := utils.JSONString(rawData, "vxcName"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("vxcName: %w", err))
	} else if present {
		req.Name = &vxcName
		fieldSet = true
	}

	if aEndInnerVLAN, present, err := utils.JSONNumber(rawData, "aEndInnerVlan"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aEndInnerVlan: %w", err))
	} else if present {
		if aEndInnerVLAN != math.Trunc(aEndInnerVLAN) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndInnerVlan must be a whole number, got %v", aEndInnerVLAN))
		}
		aEndInnerVLANInt := int(aEndInnerVLAN)
		if err := validation.ValidateVXCEndInnerVLAN(aEndInnerVLANInt); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("invalid aEndInnerVlan: %w", err))
		}
		req.AEndInnerVLAN = &aEndInnerVLANInt
		fieldSet = true
	}

	if bEndInnerVLAN, present, err := utils.JSONNumber(rawData, "bEndInnerVlan"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bEndInnerVlan: %w", err))
	} else if present {
		if bEndInnerVLAN != math.Trunc(bEndInnerVLAN) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndInnerVlan must be a whole number, got %v", bEndInnerVLAN))
		}
		bEndInnerVLANInt := int(bEndInnerVLAN)
		if err := validation.ValidateVXCEndInnerVLAN(bEndInnerVLANInt); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("invalid bEndInnerVlan: %w", err))
		}
		req.BEndInnerVLAN = &bEndInnerVLANInt
		fieldSet = true
	}

	// Handle product UIDs
	if aEndUID, present, err := utils.JSONString(rawData, "aEndUid"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aEndUid: %w", err))
	} else if present {
		req.AEndProductUID = &aEndUID
		fieldSet = true
	}

	if bEndUID, present, err := utils.JSONString(rawData, "bEndUid"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bEndUid: %w", err))
	} else if present {
		req.BEndProductUID = &bEndUID
		fieldSet = true
	}

	// Handle partner configurations - using direct map access
	if aEndPartnerConfigRaw, present, err := utils.JSONObject(rawData, "aEndPartnerConfig"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aEndPartnerConfig: %w", err))
	} else if present {
		connectType, connectTypePresent, err := utils.JSONString(aEndPartnerConfigRaw, "connectType")
		if err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndPartnerConfig.connectType: %w", err))
		}
		if !connectTypePresent || connectType == "" {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aEndPartnerConfig.connectType is required"))
		}
		if strings.ToUpper(connectType) != "VROUTER" {
			return nil, exitcodes.NewUsageError(fmt.Errorf("only VRouter partner configurations can be updated"))
		}
		aEndPartnerConfig, err := parsePartnerConfigFromMap(aEndPartnerConfigRaw)
		if err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("failed to parse A-End partner config: %w", err))
		}
		vrouterConfigA, ok := aEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig)
		if !ok {
			return nil, exitcodes.NewUsageError(fmt.Errorf("only VRouter partner configurations can be updated"))
		}
		if err := validation.ValidateVrouterPartnerConfig(vrouterConfigA); err != nil {
			return nil, exitcodes.NewUsageError(err)
		}
		req.AEndPartnerConfig = vrouterConfigA
		fieldSet = true
	}

	if bEndPartnerConfigRaw, present, err := utils.JSONObject(rawData, "bEndPartnerConfig"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bEndPartnerConfig: %w", err))
	} else if present {
		connectType, connectTypePresent, err := utils.JSONString(bEndPartnerConfigRaw, "connectType")
		if err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndPartnerConfig.connectType: %w", err))
		}
		if !connectTypePresent || connectType == "" {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bEndPartnerConfig.connectType is required"))
		}
		if strings.ToUpper(connectType) != "VROUTER" {
			return nil, exitcodes.NewUsageError(fmt.Errorf("only VRouter partner configurations can be updated"))
		}
		bEndPartnerConfig, err := parsePartnerConfigFromMap(bEndPartnerConfigRaw)
		if err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("failed to parse B-End partner config: %w", err))
		}
		vrouterConfigB, ok := bEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig)
		if !ok {
			return nil, exitcodes.NewUsageError(fmt.Errorf("only VRouter partner configurations can be updated"))
		}
		if err := validation.ValidateVrouterPartnerConfig(vrouterConfigB); err != nil {
			return nil, exitcodes.NewUsageError(err)
		}
		req.BEndPartnerConfig = vrouterConfigB
		fieldSet = true
	}

	// Handle approval and vNIC index fields from JSON
	if isApproved, present, err := utils.JSONBool(rawData, "isApproved"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("isApproved: %w", err))
	} else if present {
		req.IsApproved = &isApproved
		fieldSet = true
	}
	if aVnicIndex, present, err := utils.JSONNumber(rawData, "aVnicIndex"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("aVnicIndex: %w", err))
	} else if present {
		if aVnicIndex != math.Trunc(aVnicIndex) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("aVnicIndex must be a whole number, got %v", aVnicIndex))
		}
		idx := int(aVnicIndex)
		if err := validation.ValidateVNICIndex(idx); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("invalid aVnicIndex: %w", err))
		}
		req.AVnicIndex = &idx
		fieldSet = true
	}
	if bVnicIndex, present, err := utils.JSONNumber(rawData, "bVnicIndex"); err != nil {
		return nil, exitcodes.NewUsageError(fmt.Errorf("bVnicIndex: %w", err))
	} else if present {
		if bVnicIndex != math.Trunc(bVnicIndex) {
			return nil, exitcodes.NewUsageError(fmt.Errorf("bVnicIndex must be a whole number, got %v", bVnicIndex))
		}
		idx := int(bVnicIndex)
		if err := validation.ValidateVNICIndex(idx); err != nil {
			return nil, exitcodes.NewUsageError(fmt.Errorf("invalid bVnicIndex: %w", err))
		}
		req.BVnicIndex = &idx
		fieldSet = true
	}

	if !fieldSet {
		return nil, exitcodes.NewUsageError(fmt.Errorf("at least one field must be updated"))
	}

	// Set wait for update to true with a reasonable timeout
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	return req, nil
}
