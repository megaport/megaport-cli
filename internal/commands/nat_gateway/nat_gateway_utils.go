package nat_gateway

import (
	"context"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

var createNATGatewayFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateNATGatewayRequest) (*megaport.NATGateway, error) {
	return client.NATGatewayService.CreateNATGateway(ctx, req)
}

var listNATGatewaysFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.NATGateway, error) {
	return client.NATGatewayService.ListNATGateways(ctx)
}

var getNATGatewayFunc = func(ctx context.Context, client *megaport.Client, productUID string) (*megaport.NATGateway, error) {
	return client.NATGatewayService.GetNATGateway(ctx, productUID)
}

var updateNATGatewayFunc = func(ctx context.Context, client *megaport.Client, req *megaport.UpdateNATGatewayRequest) (*megaport.NATGateway, error) {
	return client.NATGatewayService.UpdateNATGateway(ctx, req)
}

var deleteNATGatewayFunc = func(ctx context.Context, client *megaport.Client, productUID string) error {
	return client.NATGatewayService.DeleteNATGateway(ctx, productUID)
}

var listNATGatewaySessionsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.NATGatewaySession, error) {
	return client.NATGatewayService.ListNATGatewaySessions(ctx)
}

var getNATGatewayTelemetryFunc = func(ctx context.Context, client *megaport.Client, req *megaport.GetNATGatewayTelemetryRequest) (*megaport.ServiceTelemetryResponse, error) {
	return client.NATGatewayService.GetNATGatewayTelemetry(ctx, req)
}

var validateNATGatewayOrderFunc = func(ctx context.Context, client *megaport.Client, productUID string) (*megaport.NATGatewayValidateResult, error) {
	return client.NATGatewayService.ValidateNATGatewayOrder(ctx, productUID)
}

var buyNATGatewayFunc = func(ctx context.Context, client *megaport.Client, productUID string) (*megaport.NATGatewayBuyResult, error) {
	return client.NATGatewayService.BuyNATGateway(ctx, productUID)
}

func filterNATGateways(gateways []*megaport.NATGateway, locationID int, name string) []*megaport.NATGateway {
	return utils.Filter(gateways, func(gw *megaport.NATGateway) bool {
		if gw == nil {
			return false
		}
		if locationID > 0 && gw.LocationID != locationID {
			return false
		}
		if name != "" && !strings.Contains(strings.ToLower(gw.ProductName), strings.ToLower(name)) {
			return false
		}
		return true
	})
}
