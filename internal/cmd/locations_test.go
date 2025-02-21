package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testLocations = []*megaport.Location{
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

func TestFilterLocations(t *testing.T) {
	tests := []struct {
		name     string
		filters  map[string]string
		expected int
	}{
		{
			name:     "No filters",
			filters:  map[string]string{},
			expected: 2,
		},
		{
			name:     "Filter by Metro",
			filters:  map[string]string{"metro": "Sydney"},
			expected: 1,
		},
		{
			name:     "Filter by Country",
			filters:  map[string]string{"country": "United Kingdom"},
			expected: 1,
		},
		{
			name:     "Filter by Name",
			filters:  map[string]string{"name": "London"},
			expected: 1,
		},
		{
			name:     "No match",
			filters:  map[string]string{"name": "NoMatch"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterLocations(testLocations, tt.filters)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestPrintLocations_Table(t *testing.T) {
	output := captureOutput(func() {
		printLocations(testLocations, "table")
	})

	// Table output should contain headers and both location names
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Country")
	assert.Contains(t, output, "Metro")
	assert.Contains(t, output, "Site Code")
	assert.Contains(t, output, "Status")
	assert.Contains(t, output, "Sydney")
	assert.Contains(t, output, "London")
}

func TestPrintLocations_JSON(t *testing.T) {
	output := captureOutput(func() {
		printLocations(testLocations, "json")
	})

	// JSON output should contain an array of objects
	assert.Contains(t, output, `"id":1`)
	assert.Contains(t, output, `"id":2`)
	assert.Contains(t, output, `"name":"Sydney"`)
	assert.Contains(t, output, `"name":"London"`)
	assert.Contains(t, output, `"country":"Australia"`)
	assert.Contains(t, output, `"country":"United Kingdom"`)
	assert.Contains(t, output, `"metro":"Sydney"`)
	assert.Contains(t, output, `"metro":"London"`)
	assert.Contains(t, output, `"site_code":"SYD1"`)
	assert.Contains(t, output, `"site_code":"LON1"`)
	assert.Contains(t, output, `"status":"ACTIVE"`)
}

func TestPrintLocations_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printLocations(testLocations, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
