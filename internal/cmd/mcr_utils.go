package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

	// Parse JSON into request
	req := &megaport.ModifyMCRRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Validate required fields
	if err := validateUpdateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input for updating MCR
func processFlagUpdateMCRInput(cmd *cobra.Command, mcrUID string) (*megaport.ModifyMCRRequest, error) {
	// Get required fields
	name, _ := cmd.Flags().GetString("name")

	// Get optional fields
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	termValue, _ := cmd.Flags().GetInt("term")
	marketplaceVisibilitySet := cmd.Flags().Changed("marketplace-visibility")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")

	// Build the request with required fields
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
		Name:  name,
	}

	// Set optional fields if provided
	if costCentre != "" {
		req.CostCentre = costCentre
	}

	if termValue > 0 {
		contractTermMonths := termValue
		req.ContractTermMonths = &contractTermMonths
	}

	if marketplaceVisibilitySet {
		req.MarketplaceVisibility = &marketplaceVisibility
	}

	// Validate required fields
	if err := validateUpdateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Validate MCR update request
func validateUpdateMCRRequest(req *megaport.ModifyMCRRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate term if provided
	if req.ContractTermMonths != nil {
		validTerms := []int{1, 12, 24, 36}
		validTerm := false
		for _, t := range validTerms {
			if *req.ContractTermMonths == t {
				validTerm = true
				break
			}
		}
		if !validTerm {
			return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
	}

	return nil
}

// Extract the existing interactive prompting into a separate function for MCR
func promptForMCRDetails() (*megaport.BuyMCRRequest, error) {
	name, err := prompt("Enter MCR name (required): ")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, or 36 months) (required): ")
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %v", err)
	}

	portSpeedStr, err := prompt("Enter port speed (1000, 10000, or 100000 Mbps) (required): ")
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port speed: %v", err)
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID: %v", err)
	}

	asnStr, err := prompt("Enter MCR ASN (optional): ")
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
	diversityZone, err := prompt("Enter diversity zone (optional): ")
	if err != nil {
		return nil, err
	}

	costCentre, err := prompt("Enter cost center (optional): ")
	if err != nil {
		return nil, err
	}

	promoCode, err := prompt("Enter promo code (optional): ")
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

// Extract the existing interactive prompting into a separate function for updating MCR
func promptForUpdateMCRDetails(mcrUID string) (*megaport.ModifyMCRRequest, error) {
	name, err := prompt("Enter new MCR name (required): ")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	costCentre, err := prompt("Enter new cost center (optional): ")
	if err != nil {
		return nil, err
	}

	termStr, err := prompt("Enter new term (1, 12, 24, or 36 months) (optional): ")
	if err != nil {
		return nil, err
	}

	// Build the request
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
		Name:  name,
	}

	// Set optional fields if provided
	if costCentre != "" {
		req.CostCentre = costCentre
	}

	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil {
			return nil, fmt.Errorf("invalid term: %v", err)
		}
		req.ContractTermMonths = &term
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Validate the request
	if err := validateUpdateMCRRequest(req); err != nil {
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

// Process JSON input for updating prefix filter list
func processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
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

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   tempData.Description,
		AddressFamily: tempData.AddressFamily,
		Entries:       entries,
	}

	// Validate the request
	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}

// Fix for updating prefix filter list - ensure correct types are used
func processFlagUpdatePrefixFilterListInput(cmd *cobra.Command, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	// Get fields
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

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: addressFamily,
		Entries:       entries,
	}

	// Validate required fields
	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}

// Validate update prefix filter list
func validateUpdatePrefixFilterList(prefixFilterList *megaport.MCRPrefixFilterList) error {
	if prefixFilterList.Description == "" {
		return fmt.Errorf("description is required")
	}
	if prefixFilterList.AddressFamily == "" {
		return fmt.Errorf("address family is required")
	}
	if prefixFilterList.AddressFamily != "IPv4" && prefixFilterList.AddressFamily != "IPv6" {
		return fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}
	if len(prefixFilterList.Entries) == 0 {
		return fmt.Errorf("at least one entry is required")
	}

	// Validate each entry
	for i, entry := range prefixFilterList.Entries {
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
func promptForPrefixFilterListDetails(mcrUID string) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	description, err := prompt("Enter description (required): ")
	if err != nil {
		return nil, err
	}
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	addressFamily, err := prompt("Enter address family (IPv4 or IPv6) (required): ")
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

		prefix, err := prompt("Enter prefix (e.g., 192.168.0.0/24): ")
		if err != nil {
			return nil, err
		}
		if prefix == "" {
			break
		}

		actionStr, err := prompt("Enter action (permit or deny): ")
		if err != nil {
			return nil, err
		}
		if actionStr != "permit" && actionStr != "deny" {
			return nil, fmt.Errorf("invalid action, must be permit or deny")
		}

		geStr, err := prompt("Enter GE value (optional): ")
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

		leStr, err := prompt("Enter LE value (optional): ")
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

func promptForUpdatePrefixFilterListDetails(mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
	ctx := context.Background()
	client, err := Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
	}

	fmt.Printf("Current description: %s\n", currentPrefixFilterList.Description)
	description, err := prompt("Enter new description (leave empty to keep current): ")
	if err != nil {
		return nil, err
	}
	if description == "" {
		description = currentPrefixFilterList.Description
	}

	fmt.Printf("Current address family: %s\n", currentPrefixFilterList.AddressFamily)
	addressFamily, err := prompt("Enter new address family (IPv4 or IPv6, leave empty to keep current): ")
	if err != nil {
		return nil, err
	}
	if addressFamily == "" {
		addressFamily = currentPrefixFilterList.AddressFamily
	} else if addressFamily != "IPv4" && addressFamily != "IPv6" {
		return nil, fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}

	// Initialize a zero-length slice with capacity to hold existing entries
	entries := make([]*megaport.MCRPrefixListEntry, 0, len(currentPrefixFilterList.Entries))

	modifyExisting, err := prompt("Do you want to modify existing entries? (yes/no): ")
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

			keepEntry, err := prompt(fmt.Sprintf("Keep entry %d? (yes/no): ", i+1))
			if err != nil {
				return nil, err
			}

			if strings.ToLower(keepEntry) == "yes" {
				modifyEntry, err := prompt(fmt.Sprintf("Modify entry %d? (yes/no): ", i+1))
				if err != nil {
					return nil, err
				}

				if strings.ToLower(modifyEntry) == "yes" {
					prefix, err := prompt(fmt.Sprintf("Enter new prefix for entry %d (current: %s): ", i+1, entry.Prefix))
					if err != nil {
						return nil, err
					}
					if prefix == "" {
						prefix = entry.Prefix
					}

					actionStr, err := prompt(fmt.Sprintf("Enter new action for entry %d (permit or deny, current: %s): ", i+1, entry.Action))
					if err != nil {
						return nil, err
					}
					if actionStr == "" {
						actionStr = entry.Action
					} else if actionStr != "permit" && actionStr != "deny" {
						return nil, fmt.Errorf("invalid action, must be permit or deny")
					}

					geStr, err := prompt(fmt.Sprintf("Enter new GE value for entry %d (current: %d): ", i+1, entry.Ge))
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

					leStr, err := prompt(fmt.Sprintf("Enter new LE value for entry %d (current: %d): ", i+1, entry.Le))
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

	addNew, err := prompt("Do you want to add new entries? (yes/no): ")
	if err != nil {
		return nil, err
	}

	if strings.ToLower(addNew) == "yes" {
		for {
			fmt.Println("Add a new prefix filter entry (leave prefix blank to finish):")

			prefix, err := prompt("Enter prefix (e.g., 192.168.0.0/24): ")
			if err != nil {
				return nil, err
			}
			if prefix == "" {
				break
			}

			actionStr, err := prompt("Enter action (permit or deny): ")
			if err != nil {
				return nil, err
			}
			if actionStr != "permit" && actionStr != "deny" {
				return nil, fmt.Errorf("invalid action, must be permit or deny")
			}

			geStr, err := prompt("Enter GE value (optional): ")
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

			leStr, err := prompt("Enter LE value (optional): ")
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
		AddressFamily: addressFamily,
		Entries:       entries,
	}

	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}

// PrefixFilterListOutput represents the desired fields for JSON output.
type PrefixFilterListOutput struct {
	ID            int                           `json:"id"`
	Description   string                        `json:"description"`
	AddressFamily string                        `json:"address_family"`
	Entries       []PrefixFilterListEntryOutput `json:"entries"`
}

type PrefixFilterListEntryOutput struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"`
	Le     int    `json:"le,omitempty"`
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
	output
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	ProvisioningStatus string `json:"provisioning_status"`
}

// ToMCROutput converts a *megaport.MCR to our MCROutput struct.
func ToMCROutput(mcr *megaport.MCR) (MCROutput, error) {
	if mcr == nil {
		return MCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}

	return MCROutput{
		UID:                mcr.UID,
		Name:               mcr.Name,
		LocationID:         mcr.LocationID,
		ProvisioningStatus: mcr.ProvisioningStatus,
	}, nil
}

// printMCRs prints a list of MCRs in the specified format.
func printMCRs(mcrs []*megaport.MCR, format string) error {
	outputs := make([]MCROutput, 0, len(mcrs))
	for _, mcr := range mcrs {
		output, err := ToMCROutput(mcr)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}
