package partners

import (
	"strings"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// PartnerOutput represents the desired fields for JSON output.
type PartnerOutput struct {
	output.Output `json:"-" header:"-"`
	ProductName   string `json:"product_name" header:"Name"`
	UID           string `json:"uid" header:"UID"`
	ConnectType   string `json:"connect_type" header:"Connect Type"`
	CompanyName   string `json:"company_name" header:"Company Name"`
	LocationId    int    `json:"location_id" header:"Location ID"`
	DiversityZone string `json:"diversity_zone" header:"Diversity Zone"`
	VXCPermitted  bool   `json:"vxc_permitted" header:"VXC Permitted"`
}

// ToPartnerOutput converts a PartnerMegaport to a PartnerOutput.
func ToPartnerOutput(p *megaport.PartnerMegaport) PartnerOutput {
	return PartnerOutput{
		ProductName:   p.ProductName,
		UID:           p.ProductUID,
		ConnectType:   p.ConnectType,
		CompanyName:   p.CompanyName,
		LocationId:    p.LocationId,
		DiversityZone: p.DiversityZone,
		VXCPermitted:  p.VXCPermitted,
	}
}

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

// printPartners prints the partner ports in the specified output format.
var printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string, noColor bool) error {
	// Convert partners to output format
	outputs := make([]PartnerOutput, 0, len(partners))
	for _, partner := range partners {
		outputs = append(outputs, ToPartnerOutput(partner))
	}

	// Use generic printOutput function
	return output.PrintOutput(outputs, format, noColor)
}
