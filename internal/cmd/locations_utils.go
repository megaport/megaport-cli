package cmd

import megaport "github.com/megaport/megaportgo"

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

// LocationOutput represents the desired fields for JSON output.
type LocationOutput struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Metro     string  `json:"metro"`
	SiteCode  string  `json:"site_code"`
	Market    string  `json:"market"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
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

// printLocations prints the locations in the specified output format.
func printLocations(locations []*megaport.Location, format string) error {
	outputs := make([]LocationOutput, 0, len(locations))
	for _, loc := range locations {
		outputs = append(outputs, ToLocationOutput(loc))
	}
	return printOutput(outputs, format)
}
