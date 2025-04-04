package cmd

import (
	megaport "github.com/megaport/megaportgo"
)

// filterLocations filters the provided locations based on the given filters.
func filterLocations(locations []*megaport.Location, filters map[string]string) []*megaport.Location {
	var filtered []*megaport.Location
	for _, loc := range locations {
		if metro, ok := filters["metro"]; ok && loc.Metro != metro {
			continue
		}
		if country, ok := filters["country"]; ok && loc.Country != country {
			continue
		}
		if name, ok := filters["name"]; ok && loc.Name != name {
			continue
		}
		filtered = append(filtered, loc)
	}
	return filtered
}

// LocationOutput represents the complete fields for JSON and CSV output.
type LocationOutput struct {
	ID        int     `json:"id" header:"ID"`
	Name      string  `json:"name" header:"Name"`
	Country   string  `json:"country" header:"Country"`
	Metro     string  `json:"metro" header:"Metro"`
	SiteCode  string  `json:"site_code" header:"Site Code"`
	Market    string  `json:"market" header:"Market"`
	Latitude  float64 `json:"latitude" header:"-"`  // Exclude from table output
	Longitude float64 `json:"longitude" header:"-"` // Exclude from table output
	Status    string  `json:"status" header:"Status"`
}

// ToLocationOutput converts a Location to a LocationOutput.
func ToLocationOutput(l *megaport.Location) LocationOutput {
	return LocationOutput{
		ID:        l.ID,
		Name:      l.Name,
		Country:   l.Country,
		Metro:     l.Metro,
		SiteCode:  l.SiteCode,
		Market:    l.Market,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Status:    l.Status,
	}
}

// LocationTableOutput is a compact version for table display
type LocationTableOutput struct {
	ID       int    `header:"ID"`
	Name     string `header:"Name"`
	Country  string `header:"Country"`
	Metro    string `header:"Metro"`
	SiteCode string `header:"Site Code"`
	Status   string `header:"Status"`
}

// ToLocationTableOutput converts a Location to a LocationTableOutput.
func ToLocationTableOutput(l *megaport.Location) LocationTableOutput {
	return LocationTableOutput{
		ID:       l.ID,
		Name:     l.Name,
		Country:  l.Country,
		Metro:    l.Metro,
		SiteCode: l.SiteCode,
		Status:   l.Status,
	}
}

// printLocations prints the locations in the specified output format.
func printLocations(locations []*megaport.Location, format string) error {
	// For table format, use the compact version
	if format == formatTable {
		tableOutputs := make([]LocationTableOutput, 0, len(locations))
		for _, loc := range locations {
			tableOutputs = append(tableOutputs, ToLocationTableOutput(loc))
		}
		return printOutput(tableOutputs, format)
	}

	// For JSON and CSV formats, use the full output
	outputs := make([]LocationOutput, 0, len(locations))
	for _, loc := range locations {
		outputs = append(outputs, ToLocationOutput(loc))
	}
	return printOutput(outputs, format)
}
