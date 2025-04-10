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
