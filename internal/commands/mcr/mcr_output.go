package mcr

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

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

	changes := []output.FieldChange{
		{Label: "Name", OldValue: original.Name, NewValue: updated.Name},
		{Label: "Cost Centre", OldValue: output.FormatOptionalString(original.CostCentre), NewValue: output.FormatOptionalString(updated.CostCentre)},
		{Label: "Contract Term", OldValue: fmt.Sprintf("%d months", original.ContractTermMonths), NewValue: fmt.Sprintf("%d months", updated.ContractTermMonths)},
		{Label: "Marketplace Visibility", OldValue: output.FormatBool(original.MarketplaceVisibility), NewValue: output.FormatBool(updated.MarketplaceVisibility)},
	}

	originalASN := original.Resources.VirtualRouter.ASN
	updatedASN := updated.Resources.VirtualRouter.ASN
	if originalASN != 0 || updatedASN != 0 {
		changes = append(changes, output.FieldChange{
			Label:    "ASN",
			OldValue: fmt.Sprintf("%d", originalASN),
			NewValue: fmt.Sprintf("%d", updatedASN),
		})
	}

	output.DisplayChanges(changes, noColor)
}

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

type MCRStatus struct {
	UID    string `json:"uid" header:"UID"`
	Name   string `json:"name" header:"Name"`
	Status string `json:"status" header:"Status"`
	ASN    int    `json:"asn" header:"ASN"`
	Speed  int    `json:"speed" header:"Speed"`
}
