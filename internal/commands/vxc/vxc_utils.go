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

	if partnerName == "GOOGLE" || partnerName == "ORACLE" {
		res, err = svc.LookupPartnerPorts(ctx, &megaport.LookupPartnerPortsRequest{
			Key:     key,
			Partner: partnerName,
		})
	} else {
		return "", fmt.Errorf("partner type %s does not support lookup by key", partnerName)
	}

	if err != nil {
		return "", fmt.Errorf("error looking up partner port: %v", err)
	}

	if res.ProductUID == "" {
		return "", fmt.Errorf("no partner port found for key: %s", key)
	}

	return res.ProductUID, nil
}
