//go:build integration

package ports

import (
	"context"
	"sync"

	megaport "github.com/megaport/megaportgo"
)

// integrationHookBuyResponses records BuyPortResponses keyed by request name so
// importing test packages (e.g. vxc) can recover a port's UID from the SDK
// response instead of polling ListPorts. This lives in a non-test source file
// so the hook is compiled into the ports package when it is imported under the
// integration build tag.
var integrationHookBuyResponses sync.Map // key: request.Name (string), value: *megaport.BuyPortResponse

func init() {
	base := buyPortFunc
	buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
		resp, err := base(ctx, client, req)
		if err == nil && resp != nil && req != nil && req.Name != "" {
			integrationHookBuyResponses.Store(req.Name, resp)
		}
		return resp, err
	}
}

// IntegrationBuyPortUID returns the first technical service UID captured for the
// given port name by the integration buy hook. ok is false if no response was
// recorded or it carried no UIDs. Integration-test use only.
func IntegrationBuyPortUID(name string) (uid string, ok bool) {
	v, loaded := integrationHookBuyResponses.Load(name)
	if !loaded {
		return "", false
	}
	resp, isResp := v.(*megaport.BuyPortResponse)
	if !isResp || len(resp.TechnicalServiceUIDs) == 0 {
		return "", false
	}
	return resp.TechnicalServiceUIDs[0], true
}
