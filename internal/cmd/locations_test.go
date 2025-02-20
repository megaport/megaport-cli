package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

// mockLocations is our fake data for testing
var mockLocations = []*megaport.Location{
	{
		ID:       1,
		Name:     "Sydney",
		Country:  "Australia",
		Metro:    "Sydney",
		SiteCode: "SYD1",
		Market:   "APAC",
		Status:   "ACTIVE",
	},
	{
		ID:       2,
		Name:     "London",
		Country:  "United Kingdom",
		Metro:    "London",
		SiteCode: "LON1",
		Market:   "EUROPE",
		Status:   "ACTIVE",
	},
}

// filterLocations applies filters to a list of locations.
func filterLocations(locations []*megaport.Location, filters map[string]string) []*megaport.Location {
	var filtered []*megaport.Location
	for _, loc := range locations {
		if filters["metro"] != "" && loc.Metro != filters["metro"] {
			continue
		}
		if filters["country"] != "" && loc.Country != filters["country"] {
			continue
		}
		if filters["name"] != "" && !containsCaseInsensitive(loc.Name, filters["name"]) {
			continue
		}
		filtered = append(filtered, loc)
	}
	if len(filters) == 0 {
		return locations
	}
	return filtered
}

// Simple helper to perform case-insensitive substring matching.
func containsCaseInsensitive(str, substr string) bool {
	// Your preferred approach here; this is a basic example:
	if len(substr) == 0 {
		return true
	}
	// Convert both to lower case for a simple match.
	return (len(str) >= len(substr) &&
		(len(str) > 0) &&
		(len(substr) > 0) &&
		(stringInLower(str, substr)))
}

func stringInLower(str, substr string) bool {
	s := []rune(str)
	sub := []rune(substr)
	// naive approach
	for i := 0; i <= len(s)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if toLowerRune(s[i+j]) != toLowerRune(sub[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

func TestToLocationOutput(t *testing.T) {
	loc := mockLocations[0]
	output := ToLocationOutput(loc)

	assert.Equal(t, loc.ID, output.ID)
	assert.Equal(t, loc.Name, output.Name)
	assert.Equal(t, loc.Country, output.Country)
	assert.Equal(t, loc.Metro, output.Metro)
	assert.Equal(t, loc.SiteCode, output.SiteCode)
	assert.Equal(t, loc.Market, output.Market)
	assert.Equal(t, loc.Status, output.Status)
}

func TestFilterLocations(t *testing.T) {
	tests := []struct {
		name    string
		filters map[string]string
		expect  int
	}{
		{
			name:    "No filters",
			filters: map[string]string{},
			expect:  2,
		},
		{
			name:    "Filter by metro",
			filters: map[string]string{"metro": "Sydney"},
			expect:  1,
		},
		{
			name:    "Filter by country",
			filters: map[string]string{"country": "Australia"},
			expect:  1,
		},
		{
			name:    "Filter by partial name match",
			filters: map[string]string{"name": "don"},
			expect:  1,
		},
		{
			name:    "No matches",
			filters: map[string]string{"metro": "NonExistingMetro"},
			expect:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterLocations(mockLocations, tt.filters)
			assert.Equal(t, tt.expect, len(result))
		})
	}
}
