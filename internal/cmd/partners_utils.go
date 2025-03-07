package cmd

import (
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// PartnerOutput represents the desired fields for JSON output.
type PartnerOutput struct {
	output
	ProductName   string `json:"product_name"`
	UID           string `json:"uid"`
	ConnectType   string `json:"connect_type"`
	CompanyName   string `json:"company_name"`
	LocationId    int    `json:"location_id"`
	DiversityZone string `json:"diversity_zone"`
	VXCPermitted  bool   `json:"vxc_permitted"`
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
func printPartners(partners []*megaport.PartnerMegaport, format string) error {
	// Convert partners to output format
	outputs := make([]PartnerOutput, 0, len(partners))
	for _, partner := range partners {
		outputs = append(outputs, ToPartnerOutput(partner))
	}

	// Use generic printOutput function
	return printOutput(outputs, format)
}
