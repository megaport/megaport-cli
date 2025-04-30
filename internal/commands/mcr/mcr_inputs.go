package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func processJSONMCRInput(jsonStr, jsonFile string) (*megaport.BuyMCRRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
	}

	req := &megaport.BuyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	if err := validation.ValidateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processFlagMCRInput(cmd *cobra.Command) (*megaport.BuyMCRRequest, error) {
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
			return nil, fmt.Errorf("error parsing resource tags JSON: %v", err)
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

	if err := validation.ValidateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func processJSONUpdateMCRInput(jsonStr, jsonFile string) (*megaport.ModifyMCRRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
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
		return nil, fmt.Errorf("error parsing JSON: %v", err)
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

		termValue := int64(termFloat)
		if termValue != 1 && termValue != 12 && termValue != 24 && termValue != 36 {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
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
		if term != 1 && term != 12 && term != 24 && term != 36 {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
		req.ContractTermMonths = &term
	}

	return req, nil
}

func processJSONPrefixFilterListInput(jsonStr, jsonFile string, mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
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
		return nil, fmt.Errorf("error parsing JSON: %v", err)
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
			return nil, fmt.Errorf("error parsing entries JSON: %v", err)
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

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
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
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	descriptionProvided := tempData.Description != ""
	entriesProvided := len(tempData.Entries) > 0

	if !descriptionProvided && !entriesProvided {
		return nil, fmt.Errorf("at least one field (description or entries) must be updated")
	}

	if tempData.AddressFamily != "" {
		ctx := context.Background()
		client, err := config.Login(ctx)
		if err != nil {
			return nil, err
		}

		currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
		}

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

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
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

	if cmd.Flags().Changed("address-family") {
		ctx := context.Background()
		client, err := config.Login(ctx)
		if err != nil {
			return nil, err
		}

		currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
		}

		if addressFamily != currentPrefixFilterList.AddressFamily {
			return nil, fmt.Errorf("address family cannot be changed after creation (current: %s, requested: %s)",
				currentPrefixFilterList.AddressFamily, addressFamily)
		}
	}

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
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
			return nil, fmt.Errorf("error parsing entries JSON: %v", err)
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
