package partners

import (
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func filterPartners(
	partners []*megaport.PartnerMegaport,
	productName, connectType, companyName string,
	locationID int,
	diversityZone string,
) []*megaport.PartnerMegaport {
	return utils.Filter(partners, func(partner *megaport.PartnerMegaport) bool {
		if partner == nil {
			return false
		}
		if productName != "" && !strings.Contains(strings.ToLower(partner.ProductName), strings.ToLower(productName)) {
			return false
		}
		if connectType != "" && !strings.Contains(strings.ToLower(partner.ConnectType), strings.ToLower(connectType)) {
			return false
		}
		if companyName != "" && !strings.Contains(strings.ToLower(partner.CompanyName), strings.ToLower(companyName)) {
			return false
		}
		if locationID != 0 && partner.LocationId != locationID {
			return false
		}
		if diversityZone != "" && !strings.Contains(strings.ToLower(partner.DiversityZone), strings.ToLower(diversityZone)) {
			return false
		}
		return true
	})
}
