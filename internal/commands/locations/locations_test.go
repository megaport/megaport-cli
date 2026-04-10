package locations

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
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

// RTT tests

var testRTTs = []*megaport.RoundTripTime{
	{SrcLocation: 67, DstLocation: 3, MedianRTT: 1.5},
	{SrcLocation: 67, DstLocation: 12, MedianRTT: 150.3},
	{SrcLocation: 67, DstLocation: 56, MedianRTT: 180.7},
}

// setupRTTMock configures mocks for RTT tests and returns a cleanup function
// that restores the original getRoundTripTimesFunc.
func setupRTTMock(rtts []*megaport.RoundTripTime, err error) func() {
	origFunc := getRoundTripTimesFunc
	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		client := &megaport.Client{}
		client.LocationService = &MockLocationsService{
			GetRoundTripTimesResult: rtts,
			GetRoundTripTimesErr:    err,
		}
		return client, nil
	})
	getRoundTripTimesFunc = func(ctx context.Context, client *megaport.Client, srcLocationID, year, month int) ([]*megaport.RoundTripTime, error) {
		return client.LocationService.GetRoundTripTimes(ctx, srcLocationID, year, month)
	}
	return func() { getRoundTripTimesFunc = origFunc }
}

func TestGetRoundTripTimes(t *testing.T) {
	origClientFunc := config.GetNewUnauthenticatedClientFunc()
	defer config.SetNewUnauthenticatedClientFunc(origClientFunc)

	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()
	// Fixed date for deterministic default year/month in tests.
	timeNow = func() time.Time { return time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC) }

	tests := []struct {
		name           string
		srcLocation    string
		dstLocation    string
		year           string
		month          string
		mockRTTs       []*megaport.RoundTripTime
		mockErr        error
		expectedError  string
		expectedOut    string
		notExpectedOut string
	}{
		{
			name:        "success with results",
			srcLocation: "67",
			mockRTTs:    testRTTs,
			expectedOut: "67",
		},
		{
			name:           "success with dst-location filter excludes non-matches",
			srcLocation:    "67",
			dstLocation:    "3",
			mockRTTs:       testRTTs,
			expectedOut:    "1.5",
			notExpectedOut: "150.3",
		},
		{
			name:        "success with explicit year and month",
			srcLocation: "67",
			year:        "2025",
			month:       "6",
			mockRTTs:    testRTTs,
			expectedOut: "67",
		},
		{
			name:          "API error",
			srcLocation:   "67",
			mockErr:       fmt.Errorf("API failure"),
			expectedError: "API failure",
		},
		{
			name:        "empty results",
			srcLocation: "67",
			mockRTTs:    []*megaport.RoundTripTime{},
		},
		{
			name:          "missing src-location",
			srcLocation:   "0",
			expectedError: "--src-location is required",
		},
		{
			name:          "invalid month",
			srcLocation:   "67",
			month:         "13",
			expectedError: "--month must be between 1 and 12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupRTTMock(tt.mockRTTs, tt.mockErr)
			defer cleanup()

			cmd := testutil.NewCommand("rtt", testutil.OutputAdapter(GetRoundTripTimes))
			cmd.Flags().Int("src-location", 0, "")
			cmd.Flags().Int("dst-location", 0, "")
			cmd.Flags().Int("year", 0, "")
			cmd.Flags().Int("month", 0, "")

			if tt.srcLocation != "" {
				_ = cmd.Flags().Set("src-location", tt.srcLocation)
			}
			if tt.dstLocation != "" {
				_ = cmd.Flags().Set("dst-location", tt.dstLocation)
			}
			if tt.year != "" {
				_ = cmd.Flags().Set("year", tt.year)
			}
			if tt.month != "" {
				_ = cmd.Flags().Set("month", tt.month)
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = testutil.OutputAdapter(GetRoundTripTimes)(cmd, nil)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOut != "" {
					assert.Contains(t, capturedOutput, tt.expectedOut)
				}
				if tt.notExpectedOut != "" {
					assert.NotContains(t, capturedOutput, tt.notExpectedOut)
				}
			}
		})
	}
}

func TestGetRoundTripTimesJSONOutput(t *testing.T) {
	origClientFunc := config.GetNewUnauthenticatedClientFunc()
	defer config.SetNewUnauthenticatedClientFunc(origClientFunc)

	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()

	cleanup := setupRTTMock(testRTTs, nil)
	defer cleanup()

	cmd := testutil.NewCommand("rtt", testutil.OutputAdapter(GetRoundTripTimes))
	cmd.Flags().Int("src-location", 0, "")
	cmd.Flags().Int("dst-location", 0, "")
	cmd.Flags().Int("year", 0, "")
	cmd.Flags().Int("month", 0, "")
	_ = cmd.Flags().Set("src-location", "67")
	_ = cmd.Flags().Set("output", "json")

	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = cmd.RunE(cmd, nil)
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedOutput, "src_location_id")
	assert.Contains(t, capturedOutput, "dst_location_id")
	assert.Contains(t, capturedOutput, "median_rtt_ms")
}

func TestGetRoundTripTimesDefaultsPreviousMonth(t *testing.T) {
	origClientFunc := config.GetNewUnauthenticatedClientFunc()
	defer config.SetNewUnauthenticatedClientFunc(origClientFunc)

	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()
	// Mock time to April 10, 2026 — default should be March 2026.
	timeNow = func() time.Time { return time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC) }

	var capturedYear, capturedMonth int
	origRTTFunc := getRoundTripTimesFunc
	defer func() { getRoundTripTimesFunc = origRTTFunc }()

	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		return &megaport.Client{}, nil
	})
	getRoundTripTimesFunc = func(_ context.Context, _ *megaport.Client, _ int, year, month int) ([]*megaport.RoundTripTime, error) {
		capturedYear = year
		capturedMonth = month
		return []*megaport.RoundTripTime{}, nil
	}

	cmd := testutil.NewCommand("rtt", testutil.OutputAdapter(GetRoundTripTimes))
	cmd.Flags().Int("src-location", 0, "")
	cmd.Flags().Int("dst-location", 0, "")
	cmd.Flags().Int("year", 0, "")
	cmd.Flags().Int("month", 0, "")
	_ = cmd.Flags().Set("src-location", "67")

	output.CaptureOutput(func() {
		_ = testutil.OutputAdapter(GetRoundTripTimes)(cmd, nil)
	})

	assert.Equal(t, 2026, capturedYear, "default year should be from previous month")
	assert.Equal(t, 3, capturedMonth, "default month should be previous month (March)")
}

func TestGetRoundTripTimesDefaultsJanuaryRollback(t *testing.T) {
	origClientFunc := config.GetNewUnauthenticatedClientFunc()
	defer config.SetNewUnauthenticatedClientFunc(origClientFunc)

	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()
	// Mock time to January 15, 2026 — default should be December 2025.
	timeNow = func() time.Time { return time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC) }

	var capturedYear, capturedMonth int
	origRTTFunc := getRoundTripTimesFunc
	defer func() { getRoundTripTimesFunc = origRTTFunc }()

	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		return &megaport.Client{}, nil
	})
	getRoundTripTimesFunc = func(_ context.Context, _ *megaport.Client, _ int, year, month int) ([]*megaport.RoundTripTime, error) {
		capturedYear = year
		capturedMonth = month
		return []*megaport.RoundTripTime{}, nil
	}

	cmd := testutil.NewCommand("rtt", testutil.OutputAdapter(GetRoundTripTimes))
	cmd.Flags().Int("src-location", 0, "")
	cmd.Flags().Int("dst-location", 0, "")
	cmd.Flags().Int("year", 0, "")
	cmd.Flags().Int("month", 0, "")
	_ = cmd.Flags().Set("src-location", "67")

	output.CaptureOutput(func() {
		_ = testutil.OutputAdapter(GetRoundTripTimes)(cmd, nil)
	})

	assert.Equal(t, 2025, capturedYear, "default year should roll back to 2025")
	assert.Equal(t, 12, capturedMonth, "default month should be December")
}

func TestPrintRoundTripTimes_Table(t *testing.T) {
	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = printRoundTripTimes(testRTTs, "table", noColor)
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedOutput, "SRC LOCATION ID")
	assert.Contains(t, capturedOutput, "DST LOCATION ID")
	assert.Contains(t, capturedOutput, "MEDIAN RTT (MS)")
	assert.Contains(t, capturedOutput, "67")
	assert.Contains(t, capturedOutput, "1.5")
}

func TestPrintRoundTripTimes_NilEntries(t *testing.T) {
	rtts := []*megaport.RoundTripTime{
		nil,
		{SrcLocation: 1, DstLocation: 2, MedianRTT: 5.0},
	}
	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = printRoundTripTimes(rtts, "table", noColor)
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedOutput, "5")
}

func TestPrintRoundTripTimes_Empty(t *testing.T) {
	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = printRoundTripTimes([]*megaport.RoundTripTime{}, "json", noColor)
	})

	assert.NoError(t, err)
	assert.Equal(t, "[]\n", capturedOutput)
}
