package vxc

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// VXCOutput represents the desired fields for output.
type VXCOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	AEndUID       string `json:"a_end_uid" header:"A End UID"`
	BEndUID       string `json:"b_end_uid" header:"B End UID"`
	AEndVLAN      int    `json:"a_end_vlan" header:"A End VLAN"`
	BEndVLAN      int    `json:"b_end_vlan" header:"B End VLAN"`
	RateLimit     int    `json:"rate_limit" header:"Rate Limit"`
	Status        string `json:"status" header:"Status"`
}

// ToVXCOutput converts a VXC to a VXCOutput.
func ToVXCOutput(v *megaport.VXC) (VXCOutput, error) {
	if v == nil {
		return VXCOutput{}, fmt.Errorf("invalid VXC: nil value")
	}

	aEndVLAN := v.AEndConfiguration.VLAN
	bEndVLAN := v.BEndConfiguration.VLAN
	aEndUID := v.AEndConfiguration.UID
	bEndUID := v.BEndConfiguration.UID

	status := v.ProvisioningStatus

	return VXCOutput{
		UID:       v.UID,
		Name:      v.Name,
		AEndUID:   aEndUID,
		BEndUID:   bEndUID,
		AEndVLAN:  aEndVLAN,
		BEndVLAN:  bEndVLAN,
		RateLimit: v.RateLimit,
		Status:    status,
	}, nil
}

// printVXCs prints the VXCs in the specified output format
func printVXCs(vxcs []*megaport.VXC, format string, noColor bool) error {
	if vxcs == nil {
		vxcs = []*megaport.VXC{}
	}

	outputs := make([]VXCOutput, 0, len(vxcs))
	for _, vxc := range vxcs {
		output, err := ToVXCOutput(vxc)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayVXCChanges compares the original and updated VXC and displays the differences
func displayVXCChanges(original, updated *megaport.VXC, noColor bool) {
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

	// Compare rate limit
	if original.RateLimit != updated.RateLimit {
		changesFound = true
		oldRate := output.FormatOldValue(fmt.Sprintf("%d Mbps", original.RateLimit), noColor)
		newRate := output.FormatNewValue(fmt.Sprintf("%d Mbps", updated.RateLimit), noColor)
		fmt.Printf("  • Rate Limit: %s → %s\n", oldRate, newRate)
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

	// Compare A-End VLAN - directly compare the VLAN values
	if original.AEndConfiguration.VLAN != updated.AEndConfiguration.VLAN {
		changesFound = true
		oldVlan := output.FormatOldValue(fmt.Sprintf("%d", original.AEndConfiguration.VLAN), noColor)
		newVlan := output.FormatNewValue(fmt.Sprintf("%d", updated.AEndConfiguration.VLAN), noColor)
		fmt.Printf("  • A-End VLAN: %s → %s\n", oldVlan, newVlan)
	}

	// Compare B-End VLAN - directly compare the VLAN values
	if original.BEndConfiguration.VLAN != updated.BEndConfiguration.VLAN {
		changesFound = true
		oldVlan := output.FormatOldValue(fmt.Sprintf("%d", original.BEndConfiguration.VLAN), noColor)
		newVlan := output.FormatNewValue(fmt.Sprintf("%d", updated.BEndConfiguration.VLAN), noColor)
		fmt.Printf("  • B-End VLAN: %s → %s\n", oldVlan, newVlan)
	}

	// Compare locked status
	if original.Locked != updated.Locked {
		changesFound = true
		oldLocked := "No"
		if original.Locked {
			oldLocked = "Yes"
		}
		newLocked := "No"
		if updated.Locked {
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
