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
		if productName != "" && !strings.EqualFold(partner.ProductName, productName) {
			return false
		}
		if connectType != "" && !strings.EqualFold(partner.ConnectType, connectType) {
			return false
		}
		if companyName != "" && !strings.EqualFold(partner.CompanyName, companyName) {
			return false
		}
		if locationID != 0 && partner.LocationId != locationID {
			return false
		}
		if diversityZone != "" && !strings.EqualFold(partner.DiversityZone, diversityZone) {
			return false
		}
		return true
	})
}
