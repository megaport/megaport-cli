package mve

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// MVEOutput represents the desired fields for JSON output.
type MVEOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	LocationID    int    `json:"location_id" header:"Location ID"`
	Status        string `json:"status" header:"Status"`
	Vendor        string `json:"vendor" header:"Vendor"`
	Size          string `json:"size" header:"Size"`
}

type MVEStatus struct {
	UID    string `json:"uid" header:"UID"`
	Name   string `json:"name" header:"NAME"`
	Status string `json:"status" header:"STATUS"`
	Vendor string `json:"vendor" header:"VENDOR"`
	Size   string `json:"size" header:"SIZE"`
}

// ToMVEOutput converts an MVE to an MVEOutput.
func ToMVEOutput(m *megaport.MVE) (MVEOutput, error) {
	if m == nil {
		return MVEOutput{}, fmt.Errorf("invalid MVE: nil value")
	}

	output := MVEOutput{
		UID:        m.UID,
		Name:       m.Name,
		LocationID: m.LocationID,
		Status:     m.ProvisioningStatus,
		Vendor:     m.Vendor,
		Size:       m.Size,
	}

	if m.ProvisioningStatus != "" {
		output.Status = m.ProvisioningStatus
	}

	if m.Vendor != "" {
		output.Vendor = m.Vendor
	}

	if m.Size != "" {
		output.Size = m.Size
	}

	return output, nil
}

// printMVEs prints the MVEs in the specified output format.
func printMVEs(mves []*megaport.MVE, format string, noColor bool) error {
	if mves == nil {
		mves = []*megaport.MVE{}
	}

	outputs := make([]MVEOutput, 0, len(mves))
	for _, mve := range mves {
		output, err := ToMVEOutput(mve)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayMVEChanges compares the original and updated MVE and displays the differences
func displayMVEChanges(original, updated *megaport.MVE, noColor bool) {
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

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}
