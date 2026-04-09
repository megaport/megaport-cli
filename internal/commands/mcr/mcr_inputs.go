package mcr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func processJSONMCRInput(jsonStr, jsonFile string) (*megaport.BuyMCRRequest, error) {
	var jsonData []byte
	var err error

	jsonData, err = utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.BuyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// BuyMCRRequest.AddOns is []MCRAddOn (interface) and cannot be directly
	// unmarshaled by the standard library. Read tunnelCount separately using a
	// pointer so we can distinguish "key absent" from an explicit value.
	var extras struct {
		TunnelCount *int `json:"tunnelCount"`
	}
	if err := json.Unmarshal(jsonData, &extras); err != nil {
		return nil, fmt.Errorf("failed to parse tunnelCount: %w", err)
	}
	if extras.TunnelCount != nil {
		if *extras.TunnelCount < 0 {
			return nil, fmt.Errorf("tunnelCount must be 0 or a positive value (10, 20, or 30)")
		}
		if *extras.TunnelCount > 0 {
			if err := validation.ValidateIPSecTunnelCount(*extras.TunnelCount, false); err != nil {
				return nil, err
			}
		}
		// Always include the add-on config when the key is present:
		// tunnelCount == 0 tells the API to use its default of 10 tunnels.
		req.AddOns = append(req.AddOns, &megaport.MCRAddOnIPsecConfig{
			AddOnType:   megaport.AddOnTypeIPsec,
			TunnelCount: *extras.TunnelCount,
		})
	}

	if err := validation.ValidateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processFlagMCRInput(cmd *cobra.Command) (*megaport.BuyMCRRequest, error) {
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	mcrASN, _ := cmd.Flags().GetInt("mcr-asn")

	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")

	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")
	var resourceTags map[string]string
	if resourceTagsStr != "" {
		if err := json.Unmarshal([]byte(resourceTagsStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse resource tags JSON: %w", err)
		}
	}

	req := &megaport.BuyMCRRequest{
		Name:          name,
		Term:          term,
		PortSpeed:     portSpeed,
		LocationID:    locationID,
		MCRAsn:        mcrASN,
		CostCentre:    costCentre,
		PromoCode:     promoCode,
		DiversityZone: diversityZone,
		ResourceTags:  resourceTags,
	}

	if cmd.Flags().Changed("ipsec-tunnel-count") {
		ipsecTunnelCount, _ := cmd.Flags().GetInt("ipsec-tunnel-count")
		if ipsecTunnelCount < 0 {
			return nil, fmt.Errorf("ipsec-tunnel-count must be 0 or a positive value (10, 20, or 30)")
		}
		if ipsecTunnelCount > 0 {
			if err := validation.ValidateIPSecTunnelCount(ipsecTunnelCount, false); err != nil {
				return nil, err
			}
		}
		// Always include the add-on when the flag is explicitly set:
		// ipsecTunnelCount == 0 tells the API to use its default of 10 tunnels.
		req.AddOns = append(req.AddOns, &megaport.MCRAddOnIPsecConfig{
			AddOnType:   megaport.AddOnTypeIPsec,
			TunnelCount: ipsecTunnelCount,
		})
	}

	if err := validation.ValidateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processJSONUpdateMCRInput(jsonStr, jsonFile string) (*megaport.ModifyMCRRequest, error) {
	var jsonData []byte
	var err error

	jsonData, err = utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	updateFields := []string{"name", "costCentre", "marketplaceVisibility", "contractTermMonths"}
	anyFieldUpdated := false
	for _, field := range updateFields {
		if _, ok := jsonMap[field]; ok {
			anyFieldUpdated = true
			break
		}
	}

	if !anyFieldUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	req := &megaport.ModifyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if _, nameProvided := jsonMap["name"]; nameProvided && req.Name == "" {
		return nil, fmt.Errorf("name cannot be empty if provided")
	}

	if term, provided := jsonMap["contractTermMonths"]; provided {
		if req.ContractTermMonths == nil {
			return nil, fmt.Errorf("invalid contract term: null value")
		}

		termFloat, ok := term.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid contract term type: must be a number")
		}

		// Note: fractional values (e.g. 12.5) are already rejected by the
		// json.Unmarshal into the struct's *int field above, so no need to
		// check math.Trunc here.
		termValue := int(termFloat)
		if err := validation.ValidateContractTerm(termValue); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func processFlagUpdateMCRInput(cmd *cobra.Command, mcrUID string) (*megaport.ModifyMCRRequest, error) {
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
	}

	nameSet := cmd.Flags().Changed("name")
	costCentreSet := cmd.Flags().Changed("cost-centre")
	marketplaceVisibilitySet := cmd.Flags().Changed("marketplace-visibility")
	termSet := cmd.Flags().Changed("term")

	if !nameSet && !costCentreSet && !marketplaceVisibilitySet && !termSet {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	if nameSet {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return nil, fmt.Errorf("name cannot be empty if provided")
		}
		req.Name = name
	}

	if costCentreSet {
		costCentre, _ := cmd.Flags().GetString("cost-centre")
		req.CostCentre = costCentre
	}

	if marketplaceVisibilitySet {
		marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")
		req.MarketplaceVisibility = &marketplaceVisibility
	}

	if termSet {
		term, _ := cmd.Flags().GetInt("term")
		if err := validation.ValidateContractTerm(term); err != nil {
			return nil, err
		}
		req.ContractTermMonths = &term
	}

	return req, nil
}

func processJSONPrefixFilterListInput(jsonStr, jsonFile string, mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	var jsonData []byte
	var err error

	jsonData, err = utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	var tempData struct {
		Description   string `json:"description"`
		AddressFamily string `json:"addressFamily"`
		Entries       []struct {
			Action string `json:"action"`
			Prefix string `json:"prefix"`
			Ge     *int   `json:"ge,omitempty"`
			Le     *int   `json:"le,omitempty"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(jsonData, &tempData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	entries := make([]*megaport.MCRPrefixListEntry, len(tempData.Entries))
	for i, entry := range tempData.Entries {
		var geValue int
		if entry.Ge != nil {
			geValue = *entry.Ge
		}

		var leValue int
		if entry.Le != nil {
			leValue = *entry.Le
		}

		entries[i] = &megaport.MCRPrefixListEntry{
			Action: entry.Action,
			Prefix: entry.Prefix,
			Ge:     geValue,
			Le:     leValue,
		}
	}

	req := &megaport.CreateMCRPrefixFilterListRequest{
		MCRID: mcrUID,
		PrefixFilterList: megaport.MCRPrefixFilterList{
			Description:   tempData.Description,
			AddressFamily: tempData.AddressFamily,
			Entries:       entries,
		},
	}

	if err := validation.ValidatePrefixFilterListRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processFlagPrefixFilterListInput(cmd *cobra.Command, mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	description, _ := cmd.Flags().GetString("description")
	addressFamily, _ := cmd.Flags().GetString("address-family")
	entriesJSON, _ := cmd.Flags().GetString("entries")

	var entriesData []struct {
		Action string `json:"action"`
		Prefix string `json:"prefix"`
		Ge     *int   `json:"ge,omitempty"`
		Le     *int   `json:"le,omitempty"`
	}

	if entriesJSON != "" {
		if err := json.Unmarshal([]byte(entriesJSON), &entriesData); err != nil {
			return nil, fmt.Errorf("failed to parse entries JSON: %w", err)
		}
	}

	entries := make([]*megaport.MCRPrefixListEntry, len(entriesData))
	for i, entry := range entriesData {
		var geValue int
		if entry.Ge != nil {
			geValue = *entry.Ge
		}

		var leValue int
		if entry.Le != nil {
			leValue = *entry.Le
		}

		entries[i] = &megaport.MCRPrefixListEntry{
			Action: entry.Action,
			Prefix: entry.Prefix,
			Ge:     geValue,
			Le:     leValue,
		}
	}

	req := &megaport.CreateMCRPrefixFilterListRequest{
		MCRID: mcrUID,
		PrefixFilterList: megaport.MCRPrefixFilterList{
			Description:   description,
			AddressFamily: addressFamily,
			Entries:       entries,
		},
	}

	if err := validation.ValidatePrefixFilterListRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile string, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	var jsonData []byte
	var err error

	jsonData, err = utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	var tempData struct {
		Description   string `json:"description"`
		AddressFamily string `json:"addressFamily"`
		Entries       []struct {
			Action string `json:"action"`
			Prefix string `json:"prefix"`
			Ge     *int   `json:"ge,omitempty"`
			Le     *int   `json:"le,omitempty"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(jsonData, &tempData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	descriptionProvided := tempData.Description != ""
	entriesProvided := len(tempData.Entries) > 0

	if !descriptionProvided && !entriesProvided {
		return nil, fmt.Errorf("at least one field (description or entries) must be updated")
	}

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current prefix filter list: %w", err)
	}

	if tempData.AddressFamily != "" {
		if tempData.AddressFamily != currentPrefixFilterList.AddressFamily {
			return nil, fmt.Errorf("address family cannot be changed after creation (current: %s, requested: %s)",
				currentPrefixFilterList.AddressFamily, tempData.AddressFamily)
		}
	}

	entries := make([]*megaport.MCRPrefixListEntry, len(tempData.Entries))
	for i, entry := range tempData.Entries {
		var geValue int
		if entry.Ge != nil {
			geValue = *entry.Ge
		}

		var leValue int
		if entry.Le != nil {
			leValue = *entry.Le
		}

		entries[i] = &megaport.MCRPrefixListEntry{
			Action: entry.Action,
			Prefix: entry.Prefix,
			Ge:     geValue,
			Le:     leValue,
		}
	}

	description := tempData.Description
	if !descriptionProvided {
		description = currentPrefixFilterList.Description
	}

	if !entriesProvided {
		entries = currentPrefixFilterList.Entries
	}

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: currentPrefixFilterList.AddressFamily,
		Entries:       entries,
	}

	if err := validation.ValidateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}

func processFlagUpdatePrefixFilterListInput(cmd *cobra.Command, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	descriptionProvided := cmd.Flags().Changed("description")
	entriesProvided := cmd.Flags().Changed("entries")

	if !descriptionProvided && !entriesProvided {
		return nil, fmt.Errorf("at least one field (description or entries) must be updated")
	}

	description, _ := cmd.Flags().GetString("description")
	addressFamily, _ := cmd.Flags().GetString("address-family")
	entriesJSON, _ := cmd.Flags().GetString("entries")

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current prefix filter list: %w", err)
	}

	if cmd.Flags().Changed("address-family") {
		if addressFamily != currentPrefixFilterList.AddressFamily {
			return nil, fmt.Errorf("address family cannot be changed after creation (current: %s, requested: %s)",
				currentPrefixFilterList.AddressFamily, addressFamily)
		}
	}

	if !descriptionProvided {
		description = currentPrefixFilterList.Description
	}

	entries := currentPrefixFilterList.Entries

	if entriesProvided {
		var entriesData []struct {
			Action string `json:"action"`
			Prefix string `json:"prefix"`
			Ge     *int   `json:"ge,omitempty"`
			Le     *int   `json:"le,omitempty"`
		}

		if err := json.Unmarshal([]byte(entriesJSON), &entriesData); err != nil {
			return nil, fmt.Errorf("failed to parse entries JSON: %w", err)
		}

		entries = make([]*megaport.MCRPrefixListEntry, len(entriesData))
		for i, entry := range entriesData {
			var geValue int
			if entry.Ge != nil {
				geValue = *entry.Ge
			}

			var leValue int
			if entry.Le != nil {
				leValue = *entry.Le
			}

			entries[i] = &megaport.MCRPrefixListEntry{
				Action: entry.Action,
				Prefix: entry.Prefix,
				Ge:     geValue,
				Le:     leValue,
			}
		}
	}

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: currentPrefixFilterList.AddressFamily,
		Entries:       entries,
	}

	if err := validation.ValidateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}
