package locations

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var noColor = true

var testLocations = []*megaport.LocationV3{
	{
		ID:     1,
		Name:   "Sydney",
		Metro:  "Sydney",
		Market: "APAC",
		Status: "ACTIVE",
		Address: megaport.LocationV3Address{
			Country: "Australia",
		},
		DiversityZones: &megaport.LocationV3DiversityZones{
			Red: &megaport.LocationV3DiversityZone{
				McrSpeedMbps:      []int{1000, 10000},
				MegaportSpeedMbps: []int{1, 10},
				MveAvailable:      true,
			},
		},
		DataCentre: megaport.LocationV3DataCentre{
			ID:   100,
			Name: "SYD1 Data Centre",
		},
	},
	{
		ID:     2,
		Name:   "London",
		Metro:  "London",
		Market: "EUROPE",
		Status: "ACTIVE",
		Address: megaport.LocationV3Address{
			Country: "United Kingdom",
		},
		DiversityZones: &megaport.LocationV3DiversityZones{
			Red: &megaport.LocationV3DiversityZone{
				MegaportSpeedMbps: []int{1},
			},
		},
		DataCentre: megaport.LocationV3DataCentre{
			ID:   200,
			Name: "LON1 Data Centre",
		},
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
		{
			name:     "Filter by Market",
			filters:  map[string]string{"market": "APAC"},
			expected: 1,
		},
		{
			name:     "Filter by MCR Available",
			filters:  map[string]string{"mcrAvailable": "true"},
			expected: 1,
		},
		{
			name:     "Filter by Market no match",
			filters:  map[string]string{"market": "US"},
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
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "MCR AVAILABLE")
	assert.Contains(t, output, "MVE AVAILABLE")
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
    "market": "APAC",
    "latitude": 0,
    "longitude": 0,
    "status": "ACTIVE",
    "data_centre_name": "SYD1 Data Centre",
    "data_centre_id": 100,
    "mcr_available": true,
    "mve_available": true,
    "cross_connect_available": false,
    "cross_connect_type": "",
    "ordering_message": ""
  },
  {
    "id": 2,
    "name": "London",
    "country": "United Kingdom",
    "metro": "London",
    "market": "EUROPE",
    "latitude": 0,
    "longitude": 0,
    "status": "ACTIVE",
    "data_centre_name": "LON1 Data Centre",
    "data_centre_id": 200,
    "mcr_available": false,
    "mve_available": false,
    "cross_connect_available": false,
    "cross_connect_type": "",
    "ordering_message": ""
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintLocations_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printLocations(testLocations, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `id,name,country,metro,market,latitude,longitude,status,data_centre_name,data_centre_id,mcr_available,mve_available,cross_connect_available,cross_connect_type,ordering_message
1,Sydney,Australia,Sydney,APAC,0,0,ACTIVE,SYD1 Data Centre,100,true,true,false,,
2,London,United Kingdom,London,EUROPE,0,0,ACTIVE,LON1 Data Centre,200,false,false,false,,
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
	var emptyLocations []*megaport.LocationV3
	result := filterLocations(emptyLocations, map[string]string{})
	assert.Equal(t, 0, len(result), "Expected no results for empty input")
}

func TestPrintLocations_EmptySlice(t *testing.T) {
	var emptyLocations []*megaport.LocationV3

	tableOutput := output.CaptureOutput(func() {
		err := printLocations(emptyLocations, "table", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, tableOutput, "ID")
	assert.Contains(t, tableOutput, "NAME")
	assert.Contains(t, tableOutput, "COUNTRY")
	assert.Contains(t, tableOutput, "MCR AVAILABLE")
	assert.Contains(t, tableOutput, "MVE AVAILABLE")
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
	expectedCSV := `id,name,country,metro,market,latitude,longitude,status,data_centre_name,data_centre_id,mcr_available,mve_available,cross_connect_available,cross_connect_type,ordering_message
`
	assert.Equal(t, expectedCSV, csvOutput)

	jsonOutput := output.CaptureOutput(func() {
		err := printLocations(emptyLocations, "json", noColor)
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", jsonOutput)
}
