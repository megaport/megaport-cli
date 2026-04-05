package mve

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

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
	Name   string `json:"name" header:"Name"`
	Status string `json:"status" header:"Status"`
	Vendor string `json:"vendor" header:"Vendor"`
	Size   string `json:"size" header:"Size"`
}

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

func displayMVEChanges(original, updated *megaport.MVE, noColor bool) {
	if original == nil || updated == nil {
		return
	}
	changes := []output.FieldChange{
		{Label: "Name", OldValue: original.Name, NewValue: updated.Name},
		{Label: "Cost Centre", OldValue: output.FormatOptionalString(original.CostCentre), NewValue: output.FormatOptionalString(updated.CostCentre)},
		{Label: "Contract Term", OldValue: fmt.Sprintf("%d months", original.ContractTermMonths), NewValue: fmt.Sprintf("%d months", updated.ContractTermMonths)},
		{Label: "Marketplace Visibility", OldValue: output.FormatBool(original.MarketplaceVisibility), NewValue: output.FormatBool(updated.MarketplaceVisibility)},
	}
	output.DisplayChanges(changes, noColor)
}
