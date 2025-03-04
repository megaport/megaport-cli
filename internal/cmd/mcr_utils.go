package cmd

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

// Utility functions for testing
var getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
	return client.MCRService.GetMCR(ctx, mcrUID)
}

var buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
	return client.MCRService.BuyMCR(ctx, req)
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

type MCROutput struct {
	output
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	ProvisioningStatus string `json:"provisioning_status"`
}

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

// Convert a prefix filter list to the custom output format
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
