package partners

import (
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// filterPartners applies basic in-memory filters to a list of partner ports.
func filterPartners(
	partners []*megaport.PartnerMegaport,
	productName, connectType, companyName string,
	locationID int,
	diversityZone string,
) []*megaport.PartnerMegaport {
	var filtered []*megaport.PartnerMegaport
	for _, partner := range partners {
		if productName != "" && !strings.EqualFold(partner.ProductName, productName) {
			continue
		}
		if connectType != "" && !strings.EqualFold(partner.ConnectType, connectType) {
			continue
		}
		if companyName != "" && !strings.EqualFold(partner.CompanyName, companyName) {
			continue
		}
		if locationID != 0 && partner.LocationId != locationID {
			continue
		}
		if diversityZone != "" && !strings.EqualFold(partner.DiversityZone, diversityZone) {
			continue
		}
		filtered = append(filtered, partner)
	}
	return filtered
}
