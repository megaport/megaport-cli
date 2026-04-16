package vxc

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

var deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
	err := client.VXCService.DeleteVXC(ctx, vxcUID, req)
	return err
}

var buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	return client.VXCService.BuyVXC(ctx, req)
}

var updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
	_, err := client.VXCService.UpdateVXC(ctx, vxcUID, req)
	return err
}

var getVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string) (*megaport.VXC, error) {
	return client.VXCService.GetVXC(ctx, vxcUID)
}

var getPartnerPortUID = func(ctx context.Context, svc megaport.VXCService, key, partnerName string) (string, error) {
	var res *megaport.LookupPartnerPortsResponse
	var err error

	if partnerName == "AZURE" || partnerName == "GOOGLE" || partnerName == "ORACLE" {
		res, err = svc.LookupPartnerPorts(ctx, &megaport.LookupPartnerPortsRequest{
			Key:     key,
			Partner: partnerName,
		})
	} else {
		return "", fmt.Errorf("partner type %s does not support lookup by key", partnerName)
	}

	if err != nil {
		return "", fmt.Errorf("failed to look up partner port: %w", err)
	}

	if res.ProductUID == "" {
		return "", fmt.Errorf("no partner port found for key: %s", key)
	}

	return res.ProductUID, nil
}

// resolvePartnerPortUID extracts the lookup key from a partner config and
// resolves it to a product UID. Returns ("", nil) if the config type does
// not support UID lookup.
var resolvePartnerPortUID = func(ctx context.Context, svc megaport.VXCService, partnerConfig megaport.VXCPartnerConfiguration) (string, error) {
	switch pc := partnerConfig.(type) {
	case *megaport.VXCPartnerConfigAzure:
		if pc.ServiceKey == "" {
			return "", fmt.Errorf("serviceKey is required for Azure configuration")
		}
		return getPartnerPortUID(ctx, svc, pc.ServiceKey, "AZURE")
	case *megaport.VXCPartnerConfigGoogle:
		if pc.PairingKey == "" {
			return "", fmt.Errorf("pairingKey is required for Google configuration")
		}
		return getPartnerPortUID(ctx, svc, pc.PairingKey, "GOOGLE")
	case *megaport.VXCPartnerConfigOracle:
		if pc.VirtualCircuitId == "" {
			return "", fmt.Errorf("virtualCircuitId is required for Oracle configuration")
		}
		return getPartnerPortUID(ctx, svc, pc.VirtualCircuitId, "ORACLE")
	default:
		return "", nil
	}
}

var listVXCResourceTagsFunc = func(ctx context.Context, client *megaport.Client, vxcUID string) (map[string]string, error) {
	return client.VXCService.ListVXCResourceTags(ctx, vxcUID)
}
