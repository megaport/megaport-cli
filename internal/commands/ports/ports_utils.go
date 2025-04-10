package ports

import (
	"context"
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

var getPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.Port, error) {
	return client.PortService.GetPort(ctx, portUID)
}

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

// Validate port request
func validatePortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return fmt.Errorf("port name is required")
	}
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	if req.PortSpeed != 1000 && req.PortSpeed != 10000 && req.PortSpeed != 100000 {
		return fmt.Errorf("invalid port speed, must be one of 1000, 10000, 100000")
	}
	if req.LocationId == 0 {
		return fmt.Errorf("location ID is required")
	}
	return nil
}

// Validate LAG port request
func validateLAGPortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return fmt.Errorf("port name is required")
	}
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	if req.PortSpeed != 10000 && req.PortSpeed != 100000 {
		return fmt.Errorf("invalid port speed, must be one of 10000 or 100000")
	}
	if req.LocationId == 0 {
		return fmt.Errorf("location ID is required")
	}
	if req.LagCount < 1 || req.LagCount > 8 {
		return fmt.Errorf("invalid LAG count, must be between 1 and 8")
	}
	return nil
}

// filterPorts applies filters to a list of ports
func filterPorts(ports []*megaport.Port, locationID, portSpeed int, portName string, includeInactive bool) []*megaport.Port {
	var filtered []*megaport.Port

	// Handle nil slice
	if ports == nil {
		return filtered
	}

	for _, port := range ports {
		// Skip nil ports
		if port == nil {
			continue
		}

		// Skip inactive ports if not explicitly requested
		if !includeInactive {
			if port.ProvisioningStatus == megaport.STATUS_CANCELLED ||
				port.ProvisioningStatus == megaport.STATUS_DECOMMISSIONED ||
				port.ProvisioningStatus == "DECOMMISSIONING" {
				continue
			}
		}

		// Apply other filters
		if locationID > 0 && port.LocationID != locationID {
			continue
		}
		if portSpeed > 0 && port.PortSpeed != portSpeed {
			continue
		}
		if portName != "" && !strings.Contains(strings.ToLower(port.Name), strings.ToLower(portName)) {
			continue
		}

		filtered = append(filtered, port)
	}

	return filtered
}
