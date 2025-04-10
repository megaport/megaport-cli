package mcr

import (
	"context"
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
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

// filterMCRs applies filters to a list of MCRs
func filterMCRs(mcrs []*megaport.MCR, locationID, portSpeed int, mcrName string) []*megaport.MCR {
	var filtered []*megaport.MCR

	// Handle nil slice
	if mcrs == nil {
		return filtered
	}

	for _, mcr := range mcrs {
		// Skip nil MCRs
		if mcr == nil {
			continue
		}

		// Apply filters
		if locationID > 0 && mcr.LocationID != locationID {
			continue
		}
		if portSpeed > 0 && mcr.PortSpeed != portSpeed {
			continue
		}
		if mcrName != "" && !strings.Contains(strings.ToLower(mcr.Name), strings.ToLower(mcrName)) {
			continue
		}

		filtered = append(filtered, mcr)
	}

	return filtered
}
