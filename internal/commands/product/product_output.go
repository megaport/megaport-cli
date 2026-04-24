package product

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type productOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid" header:"UID"`
	Name               string `json:"name" header:"Name"`
	Type               string `json:"type" header:"Type"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
	Speed              int    `json:"speed" header:"Speed"`
	LocationID         int    `json:"location_id" header:"Location ID"`
}

type productTypeOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Type          string `json:"type" header:"Type"`
}

func toProductOutput(p megaport.Product) (productOutput, error) {
	if p == nil {
		return productOutput{}, fmt.Errorf("invalid product: nil value")
	}
	o := productOutput{
		UID:                p.GetUID(),
		Type:               p.GetType(),
		ProvisioningStatus: p.GetProvisioningStatus(),
	}
	switch v := p.(type) {
	case *megaport.Port:
		o.Name = v.Name
		o.Speed = v.PortSpeed
		o.LocationID = v.LocationID
	case *megaport.MCR:
		o.Name = v.Name
		o.Speed = v.PortSpeed
		o.LocationID = v.LocationID
	case *megaport.MVE:
		o.Name = v.Name
		o.LocationID = v.LocationID
	}
	return o, nil
}

func printProducts(products []megaport.Product, format string, noColor bool) error {
	outputs := make([]productOutput, 0, len(products))
	for _, p := range products {
		o, err := toProductOutput(p)
		if err != nil {
			return err
		}
		outputs = append(outputs, o)
	}
	return output.PrintOutput(outputs, format, noColor)
}
