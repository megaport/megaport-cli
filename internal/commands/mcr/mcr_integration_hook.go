//go:build integration

package mcr

import (
	"context"
	"sync"

	megaport "github.com/megaport/megaportgo"
)

// integrationHookBuyResponses records BuyMCRResponses keyed by request name so
// importing test packages (e.g. vxc) can recover an MCR's UID from the SDK
// response instead of scraping stdout. This lives in a non-test source file so
// the hook is compiled into the mcr package when it is imported under the
// integration build tag.
var integrationHookBuyResponses sync.Map // key: request.Name (string), value: *megaport.BuyMCRResponse

func init() {
	base := buyMCRFunc
	buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
		resp, err := base(ctx, client, req)
		if err == nil && resp != nil && req != nil && req.Name != "" {
			integrationHookBuyResponses.Store(req.Name, resp)
		}
		return resp, err
	}
}

// IntegrationBuyMCRUID returns the technical service UID captured for the given
// MCR name by the integration buy hook. ok is false if no response was recorded
// or it carried no UID. Integration-test use only.
func IntegrationBuyMCRUID(name string) (uid string, ok bool) {
	v, loaded := integrationHookBuyResponses.Load(name)
	if !loaded {
		return "", false
	}
	resp, isResp := v.(*megaport.BuyMCRResponse)
	if !isResp || resp.TechnicalServiceUID == "" {
		return "", false
	}
	return resp.TechnicalServiceUID, true
}
