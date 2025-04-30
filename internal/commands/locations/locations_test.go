package locations

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var noColor = true

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
	var err error
	output := output.CaptureOutput(func() {
		err = printLocations(testLocations, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "COUNTRY")
	assert.Contains(t, output, "METRO")
	assert.Contains(t, output, "SITE CODE")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "Sydney")
	assert.Contains(t, output, "London")
	assert.Contains(t, output, "Australia")
	assert.Contains(t, output, "United Kingdom")
}

func TestPrintLocations_JSON(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printLocations(testLocations, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "id": 1,
    "name": "Sydney",
    "country": "Australia",
    "metro": "Sydney",
    "site_code": "SYD1",
    "market": "APAC",
    "latitude": 0,
    "longitude": 0,
    "status": "ACTIVE"
  },
  {
    "id": 2,
    "name": "London",
    "country": "United Kingdom",
    "metro": "London",
    "site_code": "LON1",
    "market": "EUROPE",
    "latitude": 0,
    "longitude": 0,
    "status": "ACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintLocations_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printLocations(testLocations, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `id,name,country,metro,site_code,market,latitude,longitude,status
1,Sydney,Australia,Sydney,SYD1,APAC,0,0,ACTIVE
2,London,United Kingdom,London,LON1,EUROPE,0,0,ACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintLocations_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printLocations(testLocations, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestFilterLocations_EmptySlice(t *testing.T) {
	var emptyLocations []*megaport.Location
	result := filterLocations(emptyLocations, map[string]string{})
	assert.Equal(t, 0, len(result), "Expected no results for empty input")
}

func TestPrintLocations_EmptySlice(t *testing.T) {
	var emptyLocations []*megaport.Location

	tableOutput := output.CaptureOutput(func() {
		err := printLocations(emptyLocations, "table", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, tableOutput, "ID")
	assert.Contains(t, tableOutput, "NAME")
	assert.Contains(t, tableOutput, "COUNTRY")
	assert.Contains(t, tableOutput, "SITE CODE")
	assert.Contains(t, tableOutput, "STATUS")
	assert.Contains(t, tableOutput, "┌")
	assert.Contains(t, tableOutput, "┐")
	assert.Contains(t, tableOutput, "└")
	assert.Contains(t, tableOutput, "┘")
	assert.Contains(t, tableOutput, "│")
	assert.Contains(t, tableOutput, "─")

	csvOutput := output.CaptureOutput(func() {
		err := printLocations(emptyLocations, "csv", noColor)
		assert.NoError(t, err)
	})
	expectedCSV := `id,name,country,metro,site_code,market,latitude,longitude,status
`
	assert.Equal(t, expectedCSV, csvOutput)

	jsonOutput := output.CaptureOutput(func() {
		err := printLocations(emptyLocations, "json", noColor)
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", jsonOutput)
}
