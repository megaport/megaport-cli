package mve

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:                "mve-1",
		Name:               "MyMVEOne",
		LocationID:         1,
		ProvisioningStatus: "LIVE",
		Vendor:             "cisco",
		Size:               "small",
	},
	{
		UID:                "mve-2",
		Name:               "AnotherMVE",
		LocationID:         2,
		ProvisioningStatus: "CONFIGURED",
		Vendor:             "palo_alto",
		Size:               "medium",
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "table", noColor)
		assert.NoError(t, err)
	})

	// Check for headers and content
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "LOCATION ID")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "VENDOR")
	assert.Contains(t, output, "SIZE")

	// Check for actual data
	assert.Contains(t, output, "mve-1")
	assert.Contains(t, output, "MyMVEOne")
	assert.Contains(t, output, "LIVE")
	assert.Contains(t, output, "cisco")
	assert.Contains(t, output, "small")

	assert.Contains(t, output, "mve-2")
	assert.Contains(t, output, "AnotherMVE")
	assert.Contains(t, output, "CONFIGURED")
	assert.Contains(t, output, "palo_alto")
	assert.Contains(t, output, "medium")

	// Check for box drawing characters
	assert.Contains(t, output, "┌")
	assert.Contains(t, output, "┐")
	assert.Contains(t, output, "└")
	assert.Contains(t, output, "┘")
	assert.Contains(t, output, "├")
	assert.Contains(t, output, "┤")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1,
    "status": "LIVE",
    "vendor": "cisco",
    "size": "small"
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2,
    "status": "CONFIGURED",
    "vendor": "palo_alto",
    "size": "medium"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,status,vendor,size
mve-1,MyMVEOne,1,LIVE,cisco,small
mve-2,AnotherMVE,2,CONFIGURED,palo_alto,medium
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printMVEs(testMVEs, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMVEs_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		mves         []*megaport.MVE
		format       string
		shouldError  bool
		validateFunc func(*testing.T, string)
	}{
		{
			name:        "nil slice",
			mves:        nil,
			format:      "table",
			shouldError: false,
			validateFunc: func(t *testing.T, output string) {
				// For table format, check for box drawing characters and headers
				assert.Contains(t, output, "UID")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "LOCATION ID")
				assert.Contains(t, output, "STATUS")
				assert.Contains(t, output, "VENDOR")
				assert.Contains(t, output, "SIZE")
				assert.Contains(t, output, "┌")
				assert.Contains(t, output, "┐")
				assert.Contains(t, output, "└")
				assert.Contains(t, output, "┘")
				assert.Contains(t, output, "│")
				assert.Contains(t, output, "─")
			},
		},
		{
			name:        "empty slice",
			mves:        []*megaport.MVE{},
			format:      "json",
			shouldError: false,
			validateFunc: func(t *testing.T, output string) {
				assert.JSONEq(t, "[]", output)
			},
		},
		{
			name: "nil mve in slice",
			mves: []*megaport.MVE{
				nil,
				{
					UID:        "mve-1",
					Name:       "TestMVE",
					LocationID: 1,
				},
			},
			format:      "table",
			shouldError: true,
			validateFunc: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
		{
			name: "zero values",
			mves: []*megaport.MVE{
				{
					UID:        "",
					Name:       "",
					LocationID: 0,
				},
			},
			format:      "csv",
			shouldError: false,
			validateFunc: func(t *testing.T, output string) {
				expected := "uid,name,location_id,status,vendor,size\n,,0,,,\n"
				assert.Equal(t, expected, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var op string
			var err error

			op = output.CaptureOutput(func() {
				err = printMVEs(tt.mves, tt.format, noColor)
			})

			if tt.shouldError {
				assert.Error(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, op)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, op)
				}
			}
		})
	}
}

func TestToMVEOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		mve           *megaport.MVE
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, MVEOutput)
	}{
		{
			name:          "nil mve",
			mve:           nil,
			shouldError:   true,
			errorContains: "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mve:  &megaport.MVE{},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Zero(t, output.LocationID)
				assert.Empty(t, output.Status)
				assert.Empty(t, output.Vendor)
				assert.Empty(t, output.Size)
			},
		},
		{
			name: "whitespace values",
			mve: &megaport.MVE{
				UID:                "   ",
				Name:               "   ",
				LocationID:         0,
				ProvisioningStatus: "   ",
				Vendor:             "   ",
				Size:               "   ",
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "   ", output.UID)
				assert.Equal(t, "   ", output.Name)
				assert.Zero(t, output.LocationID)
				assert.Equal(t, "   ", output.Status)
				assert.Equal(t, "   ", output.Vendor)
				assert.Equal(t, "   ", output.Size)
			},
		},
		{
			name: "complete values",
			mve: &megaport.MVE{
				UID:                "mve-test",
				Name:               "Test MVE",
				LocationID:         10,
				ProvisioningStatus: "LIVE",
				Vendor:             "fortinet",
				Size:               "large",
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "mve-test", output.UID)
				assert.Equal(t, "Test MVE", output.Name)
				assert.Equal(t, 10, output.LocationID)
				assert.Equal(t, "LIVE", output.Status)
				assert.Equal(t, "fortinet", output.Vendor)
				assert.Equal(t, "large", output.Size)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToMVEOutput(tt.mve)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, output)
				}
			}
		})
	}
}

func TestFilterMVEs(t *testing.T) {
	// Create test MVEs
	activeMVEs := []*megaport.MVE{
		{
			UID:                "mve-1",
			Name:               "TestMVE-1",
			LocationID:         123,
			ProvisioningStatus: "LIVE",
			Vendor:             "cisco",
		},
		{
			UID:                "mve-2",
			Name:               "TestMVE-2",
			LocationID:         456,
			ProvisioningStatus: "CONFIGURED",
			Vendor:             "fortinet",
		},
		{
			UID:                "mve-3",
			Name:               "Production-Edge",
			LocationID:         123,
			ProvisioningStatus: "LIVE",
			Vendor:             "cisco",
		},
		{
			UID:                "mve-4",
			Name:               "Staging-Edge",
			LocationID:         789,
			ProvisioningStatus: "LIVE",
			Vendor:             "versa",
		},
	}

	tests := []struct {
		name         string
		mves         []*megaport.MVE
		locationID   int
		vendor       string
		mveName      string
		expected     int      // number of MVEs after filtering
		expectedUIDs []string // specific MVE UIDs expected in result
	}{
		{
			name:         "no filters",
			mves:         activeMVEs,
			locationID:   0,
			vendor:       "",
			mveName:      "",
			expected:     4,
			expectedUIDs: []string{"mve-1", "mve-2", "mve-3", "mve-4"},
		},
		{
			name:         "filter by location ID",
			mves:         activeMVEs,
			locationID:   123,
			vendor:       "",
			mveName:      "",
			expected:     2,
			expectedUIDs: []string{"mve-1", "mve-3"},
		},
		{
			name:         "filter by vendor",
			mves:         activeMVEs,
			locationID:   0,
			vendor:       "cisco",
			mveName:      "",
			expected:     2,
			expectedUIDs: []string{"mve-1", "mve-3"},
		},
		{
			name:         "filter by vendor case insensitive",
			mves:         activeMVEs,
			locationID:   0,
			vendor:       "CiScO",
			mveName:      "",
			expected:     2,
			expectedUIDs: []string{"mve-1", "mve-3"},
		},
		{
			name:         "filter by name",
			mves:         activeMVEs,
			locationID:   0,
			vendor:       "",
			mveName:      "edge",
			expected:     2,
			expectedUIDs: []string{"mve-3", "mve-4"},
		},
		{
			name:         "filter by name and location",
			mves:         activeMVEs,
			locationID:   123,
			vendor:       "",
			mveName:      "Production",
			expected:     1,
			expectedUIDs: []string{"mve-3"},
		},
		{
			name:         "multiple filters",
			mves:         activeMVEs,
			locationID:   123,
			vendor:       "cisco",
			mveName:      "TestMVE",
			expected:     1,
			expectedUIDs: []string{"mve-1"},
		},
		{
			name:         "no matching mves",
			mves:         activeMVEs,
			locationID:   999,
			vendor:       "",
			mveName:      "",
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "nil slice",
			mves:         nil,
			locationID:   0,
			vendor:       "",
			mveName:      "",
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "empty slice",
			mves:         []*megaport.MVE{},
			locationID:   0,
			vendor:       "",
			mveName:      "",
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "slice with nil mve",
			mves:         []*megaport.MVE{nil, activeMVEs[0], nil, activeMVEs[1]},
			locationID:   0,
			vendor:       "",
			mveName:      "",
			expected:     2,
			expectedUIDs: []string{"mve-1", "mve-2"},
		},
		{
			name:         "mve with empty vendor string filtered by vendor",
			mves:         []*megaport.MVE{{UID: "mve-no-vendor", Name: "No Vendor", LocationID: 123, ProvisioningStatus: "LIVE", Vendor: ""}},
			locationID:   0,
			vendor:       "cisco",
			mveName:      "",
			expected:     0,
			expectedUIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterMVEs(tt.mves, tt.locationID, tt.vendor, tt.mveName)

			// Check the count matches
			assert.Equal(t, tt.expected, len(filtered), "Filtered MVE count should match expected")

			// Check specific UIDs if provided
			if len(tt.expectedUIDs) > 0 {
				actualUIDs := make([]string, len(filtered))
				for i, mve := range filtered {
					actualUIDs[i] = mve.UID
				}
				assert.ElementsMatch(t, tt.expectedUIDs, actualUIDs, "Filtered MVE UIDs should match expected")
			}
		})
	}
}
