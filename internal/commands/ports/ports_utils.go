package ports

import (
	"context"
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

var listPortResourceTagsFunc = func(ctx context.Context, client *megaport.Client, portUID string) (map[string]string, error) {
	return client.PortService.ListPortResourceTags(ctx, portUID)
}

func filterPorts(ports []*megaport.Port, locationID, portSpeed int, portName string, includeInactive bool) []*megaport.Port {
	var filtered []*megaport.Port
	if ports == nil {
		return filtered
	}
	for _, port := range ports {
		if port == nil {
			continue
		}
		if !includeInactive {
			if port.ProvisioningStatus == megaport.STATUS_CANCELLED ||
				port.ProvisioningStatus == megaport.STATUS_DECOMMISSIONED ||
				port.ProvisioningStatus == "DECOMMISSIONING" {
				continue
			}
		}
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
