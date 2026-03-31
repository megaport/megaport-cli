package locations

import (
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

type LocationOutput struct {
	output.Output         `json:"-" header:"-"`
	ID                    int     `json:"id" header:"ID"`
	Name                  string  `json:"name" header:"Name"`
	Country               string  `json:"country" header:"Country"`
	Metro                 string  `json:"metro" header:"Metro"`
	Market                string  `json:"market" header:"Market"`
	Latitude              float64 `json:"latitude" header:"-"`
	Longitude             float64 `json:"longitude" header:"-"`
	Status                string  `json:"status" header:"Status"`
	DataCentreName        string  `json:"data_centre_name" header:"-"`
	DataCentreID          int     `json:"data_centre_id" header:"-"`
	MCRAvailable          bool    `json:"mcr_available" header:"-"`
	MVEAvailable          bool    `json:"mve_available" header:"-"`
	CrossConnectAvailable bool    `json:"cross_connect_available" header:"-"`
	CrossConnectType      string  `json:"cross_connect_type" header:"-"`
	OrderingMessage       string  `json:"ordering_message" header:"-"`
}

func ToLocationOutput(l *megaport.LocationV3) LocationOutput {
	o := LocationOutput{
		ID:                    l.ID,
		Name:                  l.Name,
		Country:               l.Address.Country,
		Metro:                 l.Metro,
		Market:                l.Market,
		Latitude:              l.Latitude,
		Longitude:             l.Longitude,
		Status:                l.Status,
		DataCentreName:        l.GetDataCenterName(),
		DataCentreID:          l.GetDataCenterID(),
		MCRAvailable:          l.HasMCRSupport(),
		MVEAvailable:          l.HasMVESupport(),
		CrossConnectAvailable: l.HasCrossConnectSupport(),
		CrossConnectType:      l.GetCrossConnectType(),
	}
	if l.OrderingMessage != nil {
		o.OrderingMessage = *l.OrderingMessage
	}
	return o
}

type LocationTableOutput struct {
	ID           int    `header:"ID"`
	Name         string `header:"Name"`
	Country      string `header:"Country"`
	Metro        string `header:"Metro"`
	Status       string `header:"Status"`
	MCRAvailable bool   `header:"MCR Available"`
	MVEAvailable bool   `header:"MVE Available"`
}

func ToLocationTableOutput(l *megaport.LocationV3) LocationTableOutput {
	return LocationTableOutput{
		ID:           l.ID,
		Name:         l.Name,
		Country:      l.Address.Country,
		Metro:        l.Metro,
		Status:       l.Status,
		MCRAvailable: l.HasMCRSupport(),
		MVEAvailable: l.HasMVESupport(),
	}
}

type CountryOutput struct {
	output.Output `json:"-" header:"-"`
	Code          string `json:"code" header:"Code"`
	Name          string `json:"name" header:"Name"`
	Prefix        string `json:"prefix" header:"Prefix"`
	SiteCount     int    `json:"site_count" header:"Site Count"`
}

type MarketCodeOutput struct {
	output.Output `json:"-" header:"-"`
	MarketCode    string `json:"market_code" header:"Market Code"`
}

func printCountries(countries []*megaport.Country, format string, noColor bool) error {
	outputs := make([]CountryOutput, 0, len(countries))
	for _, c := range countries {
		outputs = append(outputs, CountryOutput{
			Code:      c.Code,
			Name:      c.Name,
			Prefix:    c.Prefix,
			SiteCount: c.SiteCount,
		})
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printMarketCodes(marketCodes []string, format string, noColor bool) error {
	outputs := make([]MarketCodeOutput, 0, len(marketCodes))
	for _, mc := range marketCodes {
		outputs = append(outputs, MarketCodeOutput{
			MarketCode: mc,
		})
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printLocations(locations []*megaport.LocationV3, format string, noColor bool) error {
	if format == utils.FormatTable {
		tableOutputs := make([]LocationTableOutput, 0, len(locations))
		for _, loc := range locations {
			tableOutputs = append(tableOutputs, ToLocationTableOutput(loc))
		}
		return output.PrintOutput(tableOutputs, format, noColor)
	}

	outputs := make([]LocationOutput, 0, len(locations))
	for _, loc := range locations {
		outputs = append(outputs, ToLocationOutput(loc))
	}
	return output.PrintOutput(outputs, format, noColor)
}
