package ix

import (
	"context"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

var getIXFunc = func(ctx context.Context, client *megaport.Client, ixUID string) (*megaport.IX, error) {
	return client.IXService.GetIX(ctx, ixUID)
}

var buyIXFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyIXRequest) (*megaport.BuyIXResponse, error) {
	return client.IXService.BuyIX(ctx, req)
}

var updateIXFunc = func(ctx context.Context, client *megaport.Client, ixUID string, req *megaport.UpdateIXRequest) (*megaport.IX, error) {
	return client.IXService.UpdateIX(ctx, ixUID, req)
}

var deleteIXFunc = func(ctx context.Context, client *megaport.Client, ixUID string, req *megaport.DeleteIXRequest) error {
	return client.IXService.DeleteIX(ctx, ixUID, req)
}

func filterIXs(ixs []*megaport.IX, name, networkServiceType string, asn, vlan, locationID, rateLimit int) []*megaport.IX {
	var filtered []*megaport.IX

	if ixs == nil {
		return filtered
	}

	for _, ix := range ixs {
		if ix == nil {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(ix.ProductName), strings.ToLower(name)) {
			continue
		}
		if networkServiceType != "" && !strings.Contains(strings.ToLower(ix.NetworkServiceType), strings.ToLower(networkServiceType)) {
			continue
		}
		if asn > 0 && ix.ASN != asn {
			continue
		}
		if vlan > 0 && ix.VLAN != vlan {
			continue
		}
		if locationID > 0 && ix.LocationID != locationID {
			continue
		}
		if rateLimit > 0 && ix.RateLimit != rateLimit {
			continue
		}
		filtered = append(filtered, ix)
	}

	return filtered
}
