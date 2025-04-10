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

// Function to handle VXC update API calls
var updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
	_, err := client.VXCService.UpdateVXC(ctx, vxcUID, req)
	return err
}

func getPartnerPortUID(ctx context.Context, svc megaport.VXCService, key, partner string) (string, error) {
	fmt.Println("Finding partner port...")

	partnerPortRes, err := svc.LookupPartnerPorts(ctx, &megaport.LookupPartnerPortsRequest{
		Key:     key,
		Partner: partner,
	})
	if err != nil {
		return "", fmt.Errorf("error looking up partner ports: %v", err)
	}
	return partnerPortRes.ProductUID, nil
}
