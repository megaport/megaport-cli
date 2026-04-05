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

	changes := []output.FieldChange{
		{Label: "Name", OldValue: original.Name, NewValue: updated.Name},
		{Label: "Rate Limit", OldValue: fmt.Sprintf("%d Mbps", original.RateLimit), NewValue: fmt.Sprintf("%d Mbps", updated.RateLimit)},
		{Label: "Cost Centre", OldValue: output.FormatOptionalString(original.CostCentre), NewValue: output.FormatOptionalString(updated.CostCentre)},
		{Label: "Contract Term", OldValue: fmt.Sprintf("%d months", original.ContractTermMonths), NewValue: fmt.Sprintf("%d months", updated.ContractTermMonths)},
		{Label: "A-End VLAN", OldValue: fmt.Sprintf("%d", original.AEndConfiguration.VLAN), NewValue: fmt.Sprintf("%d", updated.AEndConfiguration.VLAN)},
		{Label: "B-End VLAN", OldValue: fmt.Sprintf("%d", original.BEndConfiguration.VLAN), NewValue: fmt.Sprintf("%d", updated.BEndConfiguration.VLAN)},
		{Label: "Locked", OldValue: output.FormatBool(original.Locked), NewValue: output.FormatBool(updated.Locked)},
	}
	output.DisplayChanges(changes, noColor)
}

type VXCStatus struct {
	UID    string `json:"uid" header:"UID"`
	Name   string `json:"name" header:"Name"`
	Status string `json:"status" header:"Status"`
	Type   string `json:"type" header:"Type"`
}
