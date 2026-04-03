package ix

import (
	"context"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
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
	return utils.Filter(ixs, func(ix *megaport.IX) bool {
		if ix == nil {
			return false
		}
		if name != "" && !strings.Contains(strings.ToLower(ix.ProductName), strings.ToLower(name)) {
			return false
		}
		if networkServiceType != "" && !strings.Contains(strings.ToLower(ix.NetworkServiceType), strings.ToLower(networkServiceType)) {
			return false
		}
		if asn > 0 && ix.ASN != asn {
			return false
		}
		if vlan > 0 && ix.VLAN != vlan {
			return false
		}
		if locationID > 0 && ix.LocationID != locationID {
			return false
		}
		if rateLimit > 0 && ix.RateLimit != rateLimit {
			return false
		}
		return true
	})
}
