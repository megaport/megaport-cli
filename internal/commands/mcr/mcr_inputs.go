package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// Process JSON input (either from string or file) for buying MCR
func processJSONMCRInput(jsonStr, jsonFile string) (*megaport.BuyMCRRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Parse JSON into request
	req := &megaport.BuyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Validate required fields
	if err := validateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input for buying MCR
func processFlagMCRInput(cmd *cobra.Command) (*megaport.BuyMCRRequest, error) {
	// Get required fields
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	mcrASN, _ := cmd.Flags().GetInt("mcr-asn")

	// Get optional fields
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")

	req := &megaport.BuyMCRRequest{
		Name:          name,
		Term:          term,
		PortSpeed:     portSpeed,
		LocationID:    locationID,
		MCRAsn:        mcrASN, // Correctly spelled to match SDK
		CostCentre:    costCentre,
		PromoCode:     promoCode,
		DiversityZone: diversityZone,
	}

	// Validate required fields
	if err := validateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process JSON input (either from string or file) for updating MCR
func processJSONUpdateMCRInput(jsonStr, jsonFile string) (*megaport.ModifyMCRRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Use a map to track which fields were actually provided in the JSON
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Check that at least one field is being updated
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

	// Now parse into the actual request
	req := &megaport.ModifyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Validate name if it was provided
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

// Process flag-based input for updating MCR
func processFlagUpdateMCRInput(cmd *cobra.Command, mcrUID string) (*megaport.ModifyMCRRequest, error) {
	// Initialize request with MCR ID
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
	}

	// Check if any field is being updated
	nameSet := cmd.Flags().Changed("name")
	costCentreSet := cmd.Flags().Changed("cost-centre")
	marketplaceVisibilitySet := cmd.Flags().Changed("marketplace-visibility")
	termSet := cmd.Flags().Changed("term")

	// Make sure at least one field is being updated
	if !nameSet && !costCentreSet && !marketplaceVisibilitySet && !termSet {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	// Only add fields that were explicitly set
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
		// Validate term value before setting it
		if term != 1 && term != 12 && term != 24 && term != 36 {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
		req.ContractTermMonths = &term
	}

	return req, nil
}

// Process JSON input for creating prefix filter list
func processJSONPrefixFilterListInput(jsonStr, jsonFile string, mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Parse JSON into a temporary struct
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

	// Convert to the SDK structure
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

	// Validate the request
	if err := validatePrefixFilterListRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Fix for CreateMCRPrefixFilterListRequest - correctly structure the request
func processFlagPrefixFilterListInput(cmd *cobra.Command, mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	// Get required fields
	description, _ := cmd.Flags().GetString("description")
	addressFamily, _ := cmd.Flags().GetString("address-family")
	entriesJSON, _ := cmd.Flags().GetString("entries")

	// Parse entries from JSON string
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

	// Convert to the correct type
	entries := make([]*megaport.MCRPrefixListEntry, len(entriesData))
	for i, entry := range entriesData {
		var geValue int
		if entry.Ge != nil {
			geValue = *entry.Ge
		}

		// Similarly for Le
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

	// Validate required fields
	if err := validatePrefixFilterListRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process JSON input for updating prefix filter list
func processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile string, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Parse JSON into a temporary struct
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

	// Check if at least one field is being updated
	descriptionProvided := tempData.Description != ""
	entriesProvided := len(tempData.Entries) > 0

	if !descriptionProvided && !entriesProvided {
		return nil, fmt.Errorf("at least one field (description or entries) must be updated")
	}

	// Check if address family was provided in JSON - if so, warn that it can't be changed
	if tempData.AddressFamily != "" {
		// We need to get the current address family to validate it hasn't changed
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

	// Use the current address family instead of the one from JSON
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
	}

	// If description is not provided, keep the current one
	description := tempData.Description
	if !descriptionProvided {
		description = currentPrefixFilterList.Description
	}

	// If entries are not provided, keep the current ones
	if !entriesProvided {
		entries = currentPrefixFilterList.Entries
	}

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: currentPrefixFilterList.AddressFamily, // Always use current address family
		Entries:       entries,
	}

	// Validate the request
	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}
func processFlagUpdatePrefixFilterListInput(cmd *cobra.Command, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	// Check if required update fields are provided
	descriptionProvided := cmd.Flags().Changed("description")
	entriesProvided := cmd.Flags().Changed("entries")

	// Ensure at least one update field is provided
	if !descriptionProvided && !entriesProvided {
		return nil, fmt.Errorf("at least one field (description or entries) must be updated")
	}

	// Get fields
	description, _ := cmd.Flags().GetString("description")
	addressFamily, _ := cmd.Flags().GetString("address-family")
	entriesJSON, _ := cmd.Flags().GetString("entries")

	// Check if address family flag was set - if so, verify it hasn't changed
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

	// Get current prefix filter list to use existing values for fields that aren't being updated
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
	}

	// Use existing description if not provided
	if !descriptionProvided {
		description = currentPrefixFilterList.Description
	}

	// Use existing entries if not provided
	entries := currentPrefixFilterList.Entries

	// Parse entries from JSON string if provided
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

		// Convert to the correct type
		entries = make([]*megaport.MCRPrefixListEntry, len(entriesData))
		for i, entry := range entriesData {
			var geValue int
			if entry.Ge != nil {
				geValue = *entry.Ge
			}

			// Similarly for Le
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
		AddressFamily: currentPrefixFilterList.AddressFamily, // Always use current address family
		Entries:       entries,
	}

	// Validate required fields
	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}
