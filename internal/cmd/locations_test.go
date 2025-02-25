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
	var err error
	output := captureOutput(func() {
		err = printLocations(testLocations, "table")
		assert.NoError(t, err)
	})

	expected := `id   name     country          metro    site_code   market   latitude   longitude   status
1    Sydney   Australia        Sydney   SYD1        APAC     0          0           ACTIVE
2    London   United Kingdom   London   LON1        EUROPE   0          0           ACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintLocations_JSON(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printLocations(testLocations, "json")
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
	output := captureOutput(func() {
		err := printLocations(testLocations, "csv")
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
	output := captureOutput(func() {
		err = printLocations(testLocations, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
