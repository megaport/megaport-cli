package ports

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// PortOutput represents the desired fields for JSON output.
type PortOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid" header:"UID"`
	Name               string `json:"name" header:"Name"`
	LocationID         int    `json:"location_id" header:"LocationID"`
	PortSpeed          int    `json:"port_speed" header:"Speed"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
}

// ToPortOutput converts a *megaport.Port to our PortOutput struct.
func ToPortOutput(port *megaport.Port) (PortOutput, error) {
	if port == nil {
		return PortOutput{}, fmt.Errorf("invalid port: nil value")
	}
	return PortOutput{
		UID:                port.UID,
		Name:               port.Name,
		LocationID:         port.LocationID,
		PortSpeed:          port.PortSpeed,
		ProvisioningStatus: port.ProvisioningStatus,
	}, nil
}

func printPorts(ports []*megaport.Port, format string, noColor bool) error {
	outputs := make([]PortOutput, 0, len(ports))
	for _, port := range ports {
		output, err := ToPortOutput(port)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayPortChanges compares the original and updated Port and displays the differences
func displayPortChanges(original, updated *megaport.Port, noColor bool) {
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

	// Compare locked status
	if original.AdminLocked != updated.AdminLocked {
		changesFound = true
		oldLocked := "No"
		if original.AdminLocked {
			oldLocked = "Yes"
		}
		newLocked := "No"
		if updated.AdminLocked {
			newLocked = "Yes"
		}
		fmt.Printf("  • Locked: %s → %s\n",
			output.FormatOldValue(oldLocked, noColor),
			output.FormatNewValue(newLocked, noColor))
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}
