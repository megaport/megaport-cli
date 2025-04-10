package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// Utility functions for testing
var getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
	return client.MCRService.GetMCR(ctx, mcrUID)
}

var buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	return client.MCRService.BuyMCR(ctx, req)
}

var updateMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
	return client.MCRService.ModifyMCR(ctx, req)
}

var createMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
	return client.MCRService.CreatePrefixFilterList(ctx, req)
}

var listMCRPrefixFilterListsFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) ([]*megaport.PrefixFilterList, error) {
	return client.MCRService.ListMCRPrefixFilterLists(ctx, mcrUID)
}

var getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	return client.MCRService.GetMCRPrefixFilterList(ctx, mcrUID, prefixFilterListID)
}

var modifyMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
	return client.MCRService.ModifyMCRPrefixFilterList(ctx, mcrID, prefixFilterListID, prefixFilterList)
}

var deleteMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrID string, prefixFilterListID int) (*megaport.DeleteMCRPrefixFilterListResponse, error) {
	return client.MCRService.DeleteMCRPrefixFilterList(ctx, mcrID, prefixFilterListID)
}

var deleteMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.DeleteMCRRequest) (*megaport.DeleteMCRResponse, error) {
	return client.MCRService.DeleteMCR(ctx, req)
}

var restoreMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.RestoreMCRResponse, error) {
	return client.MCRService.RestoreMCR(ctx, mcrUID)
}

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

// Validate MCR request
func validateMCRRequest(req *megaport.BuyMCRRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Term == 0 {
		return fmt.Errorf("term is required")
	}

	// Then validate that term is one of the allowed values
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	if req.PortSpeed == 0 {
		return fmt.Errorf("port speed is required")
	}

	if req.LocationID == 0 {
		return fmt.Errorf("location ID is required")
	}

	return nil
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

// Extract the existing interactive prompting into a separate function for updating MCR
func promptForUpdateMCRDetails(mcrUID string, noColor bool) (*megaport.ModifyMCRRequest, error) {
	// Initialize request with MCR ID
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
	}

	// Track if any field is updated
	fieldsUpdated := false

	// Prompt for name (can be skipped with empty input)
	namePrompt := "Enter new MCR name (leave empty to skip): "
	name, err := utils.Prompt(namePrompt, noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = name
		fieldsUpdated = true
	}

	// Prompt for cost centre (optional)
	costCentrePrompt := "Enter new cost centre (leave empty to skip): "
	costCentre, err := utils.Prompt(costCentrePrompt, noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = costCentre
		fieldsUpdated = true
	}

	// Prompt for marketplace visibility
	marketplaceVisibilityPrompt := "Update marketplace visibility? (yes/no, leave empty to skip): "
	marketplaceVisibilityStr, err := utils.Prompt(marketplaceVisibilityPrompt, noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(marketplaceVisibilityStr) == "yes" {
		visibilityValuePrompt := "Enter marketplace visibility (true or false): "
		visibilityValue, err := utils.Prompt(visibilityValuePrompt, noColor)
		if err != nil {
			return nil, err
		}

		marketplaceVisibility := strings.ToLower(visibilityValue) == "true"
		req.MarketplaceVisibility = &marketplaceVisibility
		fieldsUpdated = true
	}

	// Prompt for term (optional)
	termPrompt := "Enter new term (1, 12, 24, or 36 months, leave empty to skip): "
	termStr, err := utils.Prompt(termPrompt, noColor)
	if err != nil {
		return nil, err
	}
	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil {
			return nil, fmt.Errorf("invalid term: %v", err)
		}

		// Validate term value
		if term != 1 && term != 12 && term != 24 && term != 36 {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}

		req.ContractTermMonths = &term
		fieldsUpdated = true
	}

	// Make sure at least one field is being updated
	if !fieldsUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	return req, nil
}

// Extract the existing interactive prompting into a separate function for MCR
func promptForMCRDetails(noColor bool) (*megaport.BuyMCRRequest, error) {
	name, err := utils.Prompt("Enter MCR name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	termStr, err := utils.Prompt("Enter term (1, 12, 24, or 36 months) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %v", err)
	}

	portSpeedStr, err := utils.Prompt("Enter port speed (1000, 10000, or 100000 Mbps) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port speed: %v", err)
	}

	locationIDStr, err := utils.Prompt("Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID: %v", err)
	}

	asnStr, err := utils.Prompt("Enter MCR ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	var asn int
	if asnStr != "" {
		asnValue, err := strconv.Atoi(asnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ASN: %v", err)
		}
		asn = asnValue
	}

	// Optional fields
	diversityZone, err := utils.Prompt("Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	costCentre, err := utils.Prompt("Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	promoCode, err := utils.Prompt("Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	req := &megaport.BuyMCRRequest{
		Name:          name,
		Term:          term,
		PortSpeed:     portSpeed,
		LocationID:    locationID,
		MCRAsn:        asn,
		DiversityZone: diversityZone,
		CostCentre:    costCentre,
		PromoCode:     promoCode,
	}

	// Validate the request
	if err := validateMCRRequest(req); err != nil {
		return nil, err
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

// Update the validation function for prefix filter list requests
func validatePrefixFilterListRequest(req *megaport.CreateMCRPrefixFilterListRequest) error {
	if req.PrefixFilterList.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.PrefixFilterList.AddressFamily == "" {
		return fmt.Errorf("address family is required")
	}
	if req.PrefixFilterList.AddressFamily != "IPv4" && req.PrefixFilterList.AddressFamily != "IPv6" {
		return fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}
	if len(req.PrefixFilterList.Entries) == 0 {
		return fmt.Errorf("at least one entry is required")
	}

	// Validate each entry
	for i, entry := range req.PrefixFilterList.Entries {
		if entry.Prefix == "" {
			return fmt.Errorf("entry %d: prefix is required", i+1)
		}
		if entry.Action != "permit" && entry.Action != "deny" {
			return fmt.Errorf("entry %d: invalid action, must be permit or deny", i+1)
		}
	}

	return nil
}

// Also fix promptForPrefixFilterListDetails to return the correct structure
func promptForPrefixFilterListDetails(mcrUID string, noColor bool) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	description, err := utils.Prompt("Enter description (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	addressFamily, err := utils.Prompt("Enter address family (IPv4 or IPv6) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if addressFamily != "IPv4" && addressFamily != "IPv6" {
		return nil, fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}

	// Prompt for entries
	entries := []*megaport.MCRPrefixListEntry{}
	for {
		fmt.Println("Add a new prefix filter entry (leave prefix blank to finish):")

		prefix, err := utils.Prompt("Enter prefix (e.g., 192.168.0.0/24): ", noColor)
		if err != nil {
			return nil, err
		}
		if prefix == "" {
			break
		}

		actionStr, err := utils.Prompt("Enter action (permit or deny): ", noColor)
		if err != nil {
			return nil, err
		}
		if actionStr != "permit" && actionStr != "deny" {
			return nil, fmt.Errorf("invalid action, must be permit or deny")
		}

		geStr, err := utils.Prompt("Enter GE value (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		var ge int
		if geStr != "" {
			geVal, err := strconv.Atoi(geStr)
			if err != nil {
				return nil, fmt.Errorf("invalid GE value, must be a number")
			}
			ge = geVal
		}

		leStr, err := utils.Prompt("Enter LE value (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		var le int
		if leStr != "" {
			leVal, err := strconv.Atoi(leStr)
			if err != nil {
				return nil, fmt.Errorf("invalid LE value, must be a number")
			}
			le = leVal
		}

		entries = append(entries, &megaport.MCRPrefixListEntry{
			Prefix: prefix,
			Action: actionStr,
			Ge:     ge,
			Le:     le,
		})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("at least one entry is required")
	}

	req := &megaport.CreateMCRPrefixFilterListRequest{
		MCRID: mcrUID,
		PrefixFilterList: megaport.MCRPrefixFilterList{
			Description:   description,
			AddressFamily: addressFamily,
			Entries:       entries,
		},
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

func validateUpdatePrefixFilterList(prefixFilterList *megaport.MCRPrefixFilterList) error {
	// If entries are provided, validate them
	if len(prefixFilterList.Entries) > 0 {
		// Validate each entry
		for i, entry := range prefixFilterList.Entries {
			if entry.Prefix == "" {
				return fmt.Errorf("entry %d: prefix is required", i+1)
			}
			if entry.Action != "permit" && entry.Action != "deny" {
				return fmt.Errorf("entry %d: invalid action, must be permit or deny", i+1)
			}
		}
	}

	return nil
}

// Update the interactive prompting function to not allow changing address family
func promptForUpdatePrefixFilterListDetails(mcrUID string, prefixFilterListID int, noColor bool) (*megaport.MCRPrefixFilterList, error) {
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
	}

	fmt.Printf("Current description: %s\n", currentPrefixFilterList.Description)
	description, err := utils.Prompt("Enter new description (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if description == "" {
		description = currentPrefixFilterList.Description
	}

	// Just display the address family but don't allow changing it
	fmt.Printf("Address family: %s (cannot be changed after creation)\n", currentPrefixFilterList.AddressFamily)
	addressFamily := currentPrefixFilterList.AddressFamily

	// Initialize a zero-length slice with capacity to hold existing entries
	entries := make([]*megaport.MCRPrefixListEntry, 0, len(currentPrefixFilterList.Entries))

	modifyExisting, err := utils.Prompt("Do you want to modify existing entries? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(modifyExisting) != "yes" {
		// Just keep existing entries as is
		entries = append(entries, currentPrefixFilterList.Entries...)
	} else {
		for i, entry := range currentPrefixFilterList.Entries {
			fmt.Printf("Entry %d - Current: Action: %s, Prefix: %s, GE: %d, LE: %d\n",
				i+1, entry.Action, entry.Prefix, entry.Ge, entry.Le)

			keepEntry, err := utils.Prompt(fmt.Sprintf("Keep entry %d? (yes/no): ", i+1), noColor)
			if err != nil {
				return nil, err
			}

			if strings.ToLower(keepEntry) == "yes" {
				modifyEntry, err := utils.Prompt(fmt.Sprintf("Modify entry %d? (yes/no): ", i+1), noColor)
				if err != nil {
					return nil, err
				}

				if strings.ToLower(modifyEntry) == "yes" {
					prefix, err := utils.Prompt(fmt.Sprintf("Enter new prefix for entry %d (current: %s): ", i+1, entry.Prefix), noColor)
					if err != nil {
						return nil, err
					}
					if prefix == "" {
						prefix = entry.Prefix
					}

					actionStr, err := utils.Prompt(fmt.Sprintf("Enter new action for entry %d (permit or deny, current: %s): ", i+1, entry.Action), noColor)
					if err != nil {
						return nil, err
					}
					if actionStr == "" {
						actionStr = entry.Action
					} else if actionStr != "permit" && actionStr != "deny" {
						return nil, fmt.Errorf("invalid action, must be permit or deny")
					}

					geStr, err := utils.Prompt(fmt.Sprintf("Enter new GE value for entry %d (current: %d): ", i+1, entry.Ge), noColor)
					if err != nil {
						return nil, err
					}
					ge := entry.Ge
					if geStr != "" {
						geVal, err := strconv.Atoi(geStr)
						if err != nil {
							return nil, fmt.Errorf("invalid GE value, must be a number")
						}
						ge = geVal
					}

					leStr, err := utils.Prompt(fmt.Sprintf("Enter new LE value for entry %d (current: %d): ", i+1, entry.Le), noColor)
					if err != nil {
						return nil, err
					}
					le := entry.Le
					if leStr != "" {
						leVal, err := strconv.Atoi(leStr)
						if err != nil {
							return nil, fmt.Errorf("invalid LE value, must be a number")
						}
						le = leVal
					}

					entries = append(entries, &megaport.MCRPrefixListEntry{
						Prefix: prefix,
						Action: actionStr,
						Ge:     ge,
						Le:     le,
					})
				} else {
					entries = append(entries, entry)
				}
			}
		}
	}

	addNew, err := utils.Prompt("Do you want to add new entries? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(addNew) == "yes" {
		for {
			fmt.Println("Add a new prefix filter entry (leave prefix blank to finish):")

			prefix, err := utils.Prompt("Enter prefix (e.g., 192.168.0.0/24): ", noColor)
			if err != nil {
				return nil, err
			}
			if prefix == "" {
				break
			}

			actionStr, err := utils.Prompt("Enter action (permit or deny): ", noColor)
			if err != nil {
				return nil, err
			}
			if actionStr != "permit" && actionStr != "deny" {
				return nil, fmt.Errorf("invalid action, must be permit or deny")
			}

			geStr, err := utils.Prompt("Enter GE value (optional): ", noColor)
			if err != nil {
				return nil, err
			}
			var ge int
			if geStr != "" {
				geVal, err := strconv.Atoi(geStr)
				if err != nil {
					return nil, fmt.Errorf("invalid GE value, must be a number")
				}
				ge = geVal
			}

			leStr, err := utils.Prompt("Enter LE value (optional): ", noColor)
			if err != nil {
				return nil, err
			}
			var le int
			if leStr != "" {
				leVal, err := strconv.Atoi(leStr)
				if err != nil {
					return nil, fmt.Errorf("invalid LE value, must be a number")
				}
				le = leVal
			}

			entries = append(entries, &megaport.MCRPrefixListEntry{
				Prefix: prefix,
				Action: actionStr,
				Ge:     ge,
				Le:     le,
			})
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("at least one entry is required")
	}

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: addressFamily, // Always use current address family
		Entries:       entries,
	}

	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}

// PrefixFilterListOutput represents the desired fields for JSON output.
type PrefixFilterListOutput struct {
	output.Output `json:"-" header:"-"`
	ID            int                           `json:"id"`
	Description   string                        `json:"description"`
	AddressFamily string                        `json:"address_family"`
	Entries       []PrefixFilterListEntryOutput `json:"entries"`
}

type PrefixFilterListEntryOutput struct {
	output.Output `json:"-" header:"-"`
	Action        string `json:"action"`
	Prefix        string `json:"prefix"`
	Ge            int    `json:"ge,omitempty"`
	Le            int    `json:"le,omitempty"`
}

// ToPrefixFilterListOutput converts a *megaport.MCRPrefixFilterList to our PrefixFilterListOutput struct.
func ToPrefixFilterListOutput(prefixFilterList *megaport.MCRPrefixFilterList) (PrefixFilterListOutput, error) {
	if prefixFilterList == nil {
		return PrefixFilterListOutput{}, fmt.Errorf("invalid prefix filter list: nil value")
	}

	entries := make([]PrefixFilterListEntryOutput, len(prefixFilterList.Entries))
	for i, entry := range prefixFilterList.Entries {
		entries[i] = PrefixFilterListEntryOutput{
			Action: entry.Action,
			Prefix: entry.Prefix,
			Ge:     entry.Ge,
			Le:     entry.Le,
		}
	}

	return PrefixFilterListOutput{
		ID:            prefixFilterList.ID,
		Description:   prefixFilterList.Description,
		AddressFamily: prefixFilterList.AddressFamily,
		Entries:       entries,
	}, nil
}

// MCROutput represents the desired fields for JSON output of MCR details.
type MCROutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid" header:"UID"`
	Name               string `json:"name" header:"Name"`
	LocationID         int    `json:"location_id" header:"Location ID"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
	ASN                int    `json:"asn" header:"ASN"`
	Speed              int    `json:"speed" header:"Speed"`
}

// ToMCROutput converts a *megaport.MCR to our MCROutput struct.
func ToMCROutput(mcr *megaport.MCR) (MCROutput, error) {
	if mcr == nil {
		return MCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}

	output := MCROutput{
		UID:                mcr.UID,
		Name:               mcr.Name,
		LocationID:         mcr.LocationID,
		ProvisioningStatus: mcr.ProvisioningStatus,
		Speed:              mcr.PortSpeed,
	}

	output.ASN = mcr.Resources.VirtualRouter.ASN

	return output, nil
}

// printMCRs prints a list of MCRs in the specified format.
func printMCRs(mcrs []*megaport.MCR, format string, noColor bool) error {
	outputs := make([]MCROutput, 0, len(mcrs))
	for _, mcr := range mcrs {
		output, err := ToMCROutput(mcr)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayMCRChanges compares the original and updated MCR and displays the differences
func displayMCRChanges(original, updated *megaport.MCR, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	fmt.Println() // Empty line before changes
	output.PrintInfo("Changes applied:", noColor)

	// Track if any changes were found
	changesFound := false

	// Compare name
	if original.Name != updated.Name {
		changesFound = true
		oldName := output.FormatOldValue(original.Name, noColor)
		newName := output.FormatNewValue(updated.Name, noColor)
		fmt.Printf("  • Name: %s → %s\n", oldName, newName)
	}

	// Compare cost centre
	if original.CostCentre != updated.CostCentre {
		changesFound = true
		oldCostCentre := original.CostCentre
		if oldCostCentre == "" {
			oldCostCentre = "(none)"
		}
		newCostCentre := updated.CostCentre
		if newCostCentre == "" {
			newCostCentre = "(none)"
		}
		fmt.Printf("  • Cost Centre: %s → %s\n",
			output.FormatOldValue(oldCostCentre, noColor),
			output.FormatNewValue(newCostCentre, noColor))
	}

	// Compare contract term
	if original.ContractTermMonths != updated.ContractTermMonths {
		changesFound = true
		oldTerm := output.FormatOldValue(fmt.Sprintf("%d months", original.ContractTermMonths), noColor)
		newTerm := output.FormatNewValue(fmt.Sprintf("%d months", updated.ContractTermMonths), noColor)
		fmt.Printf("  • Contract Term: %s → %s\n", oldTerm, newTerm)
	}

	// Compare marketplace visibility
	if original.MarketplaceVisibility != updated.MarketplaceVisibility {
		changesFound = true
		oldVisibility := "No"
		if original.MarketplaceVisibility {
			oldVisibility = "Yes"
		}
		newVisibility := "No"
		if updated.MarketplaceVisibility {
			newVisibility = "Yes"
		}
		fmt.Printf("  • Marketplace Visibility: %s → %s\n",
			output.FormatOldValue(oldVisibility, noColor),
			output.FormatNewValue(newVisibility, noColor))
	}

	originalASN := original.Resources.VirtualRouter.ASN
	updatedASN := updated.Resources.VirtualRouter.ASN

	// Compare ASN if it changed
	if originalASN != updatedASN && (originalASN != 0 || updatedASN != 0) {
		changesFound = true
		oldASN := output.FormatOldValue(fmt.Sprintf("%d", originalASN), noColor)
		newASN := output.FormatNewValue(fmt.Sprintf("%d", updatedASN), noColor)
		fmt.Printf("  • ASN: %s → %s\n", oldASN, newASN)
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}
