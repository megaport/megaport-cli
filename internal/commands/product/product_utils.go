package product

import (
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func filterProducts(products []megaport.Product, includeInactive bool) []megaport.Product {
	var filtered []megaport.Product
	for _, p := range products {
		if p == nil {
			continue
		}
		if !includeInactive {
			status := p.GetProvisioningStatus()
			if status == megaport.STATUS_CANCELLED ||
				status == megaport.STATUS_DECOMMISSIONED ||
				status == utils.StatusDecommissioning {
				continue
			}
		}
		filtered = append(filtered, p)
	}
	return filtered
}
