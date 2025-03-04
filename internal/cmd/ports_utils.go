package cmd

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

var updatePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return client.PortService.ModifyPort(ctx, req)
}

var deletePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return client.PortService.DeletePort(ctx, req)
}

var restorePortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.RestorePortResponse, error) {
	return client.PortService.RestorePort(ctx, portUID)
}

var lockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.LockPortResponse, error) {
	return client.PortService.LockPort(ctx, portUID)
}

var unlockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.UnlockPortResponse, error) {
	return client.PortService.UnlockPort(ctx, portUID)
}

var checkPortVLANAvailabilityFunc = func(ctx context.Context, client *megaport.Client, portUID string, vlan int) (bool, error) {
	return client.PortService.CheckPortVLANAvailability(ctx, portUID, vlan)
}

var buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return client.PortService.BuyPort(ctx, req)
}

// filterPorts filters the provided ports based on the given filters.
func filterPorts(ports []*megaport.Port, locationID int, portSpeed int, portName string) []*megaport.Port {
	if ports == nil {
		return []*megaport.Port{}
	}

	filteredPorts := make([]*megaport.Port, 0)

	for _, port := range ports {
		if port == nil {
			continue
		}

		// Apply location ID filter
		if locationID != 0 && port.LocationID != locationID {
			continue
		}

		// Apply port speed filter
		if portSpeed != 0 && port.PortSpeed != portSpeed {
			continue
		}

		// Apply port name filter
		if portName != "" && port.Name != portName {
			continue
		}

		// Port passed all filters
		filteredPorts = append(filteredPorts, port)
	}

	return filteredPorts
}

// PortOutput represents the desired fields for JSON output.
type PortOutput struct {
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	PortSpeed          int    `json:"port_speed"`
	ProvisioningStatus string `json:"provisioning_status"`
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

func printPorts(ports []*megaport.Port, format string) error {
	outputs := make([]PortOutput, 0, len(ports))
	for _, port := range ports {
		output, err := ToPortOutput(port)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}
