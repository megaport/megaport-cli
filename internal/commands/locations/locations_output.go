package locations

import (
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

type LocationOutput struct {
	output.Output `json:"-" header:"-"`
	ID            int     `json:"id" header:"ID"`
	Name          string  `json:"name" header:"Name"`
	Country       string  `json:"country" header:"Country"`
	Metro         string  `json:"metro" header:"Metro"`
	SiteCode      string  `json:"site_code" header:"Site Code"` // Note: Site code deprecated in v3 API
	Market        string  `json:"market" header:"Market"`
	Latitude      float64 `json:"latitude" header:"-"`
	Longitude     float64 `json:"longitude" header:"-"`
	Status        string  `json:"status" header:"Status"`
}

func ToLocationOutput(l *megaport.Location) LocationOutput {
	return LocationOutput{
		ID:        l.ID,
		Name:      l.Name,
		Country:   l.Country,
		Metro:     l.Metro,
		SiteCode:  l.SiteCode, // Will be empty for v3-sourced data
		Market:    l.Market,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Status:    l.Status,
	}
}

type LocationTableOutput struct {
	ID       int    `header:"ID"`
	Name     string `header:"Name"`
	Country  string `header:"Country"`
	Metro    string `header:"Metro"`
	SiteCode string `header:"Site Code"` // Note: Site code deprecated in v3 API
	Status   string `header:"Status"`
}

func ToLocationTableOutput(l *megaport.Location) LocationTableOutput {
	return LocationTableOutput{
		ID:       l.ID,
		Name:     l.Name,
		Country:  l.Country,
		Metro:    l.Metro,
		SiteCode: l.SiteCode, // Will be empty for v3-sourced data
		Status:   l.Status,
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

func printLocations(locations []*megaport.Location, format string, noColor bool) error {
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
