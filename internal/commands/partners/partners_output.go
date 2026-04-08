package partners

import (
	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type partnerOutput struct {
	output.Output `json:"-" header:"-"`
	ProductName   string `json:"product_name" header:"Name"`
	UID           string `json:"uid" header:"UID"`
	ConnectType   string `json:"connect_type" header:"Connect Type"`
	CompanyName   string `json:"company_name" header:"Company Name"`
	LocationId    int    `json:"location_id" header:"Location ID"`
	DiversityZone string `json:"diversity_zone" header:"Diversity Zone"`
	VXCPermitted  bool   `json:"vxc_permitted" header:"VXC Permitted"`
}

func toPartnerOutput(p *megaport.PartnerMegaport) partnerOutput {
	return partnerOutput{
		ProductName:   p.ProductName,
		UID:           p.ProductUID,
		ConnectType:   p.ConnectType,
		CompanyName:   p.CompanyName,
		LocationId:    p.LocationId,
		DiversityZone: p.DiversityZone,
		VXCPermitted:  p.VXCPermitted,
	}
}

var printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string, noColor bool) error {
	outputs := make([]partnerOutput, 0, len(partners))
	for _, partner := range partners {
		outputs = append(outputs, toPartnerOutput(partner))
	}
	return output.PrintOutput(outputs, format, noColor)
}
