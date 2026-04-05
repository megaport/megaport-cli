package ports

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type PortOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid" header:"UID"`
	Name               string `json:"name" header:"Name"`
	LocationID         int    `json:"location_id" header:"Location ID"`
	PortSpeed          int    `json:"port_speed" header:"Speed"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
}

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

type PortStatus struct {
	UID    string `json:"uid" header:"UID"`
	Name   string `json:"name" header:"Name"`
	Status string `json:"status" header:"Status"`
	Type   string `json:"type" header:"Type"`
	Speed  int    `json:"speed" header:"Speed"`
}

func displayPortChanges(original, updated *megaport.Port, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	changes := []output.FieldChange{
		{Label: "Name", OldValue: original.Name, NewValue: updated.Name},
		{Label: "Cost Centre", OldValue: output.FormatOptionalString(original.CostCentre), NewValue: output.FormatOptionalString(updated.CostCentre)},
		{Label: "Contract Term", OldValue: fmt.Sprintf("%d months", original.ContractTermMonths), NewValue: fmt.Sprintf("%d months", updated.ContractTermMonths)},
		{Label: "Marketplace Visibility", OldValue: output.FormatBool(original.MarketplaceVisibility), NewValue: output.FormatBool(updated.MarketplaceVisibility)},
		{Label: "Locked", OldValue: output.FormatBool(original.AdminLocked), NewValue: output.FormatBool(updated.AdminLocked)},
	}

	output.DisplayChanges(changes, noColor)
}
