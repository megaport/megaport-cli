package cmd

import (
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:                "mve-1",
		Name:               "MyMVEOne",
		LocationID:         1,
		LocationDetails:    &megaport.ProductLocationDetails{Name: "Sydney"},
		Vendor:             "Cisco",
		Size:               "MEDIUM",
		ProvisioningStatus: "LIVE",
		CreateDate:         &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		ContractTermMonths: 12,
	},
	{
		UID:                "mve-2",
		Name:               "AnotherMVE",
		LocationID:         2,
		LocationDetails:    &megaport.ProductLocationDetails{Name: "Melbourne"},
		Vendor:             "Palo Alto",
		Size:               "LARGE",
		ProvisioningStatus: "CONFIGURING",
		CreateDate:         &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		ContractTermMonths: 24,
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "table")
		assert.NoError(t, err)
	})

	// Fix spacing to match actual output
	expected := `uid     name         location_id   location_name   vendor      size     provisioning_status   create_date   contract_term_months
mve-1   MyMVEOne     1             Sydney          Cisco       MEDIUM   LIVE                  2023-01-01    12
mve-2   AnotherMVE   2             Melbourne       Palo Alto   LARGE    CONFIGURING           2023-02-01    24
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1,
    "location_name": "Sydney",
    "vendor": "Cisco",
    "size": "MEDIUM",
    "provisioning_status": "LIVE",
    "create_date": "2023-01-01",
    "contract_term_months": 12
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2,
    "location_name": "Melbourne",
    "vendor": "Palo Alto",
    "size": "LARGE",
    "provisioning_status": "CONFIGURING",
    "create_date": "2023-02-01",
    "contract_term_months": 24
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,location_name,vendor,size,provisioning_status,create_date,contract_term_months
mve-1,MyMVEOne,1,Sydney,Cisco,MEDIUM,LIVE,2023-01-01,12
mve-2,AnotherMVE,2,Melbourne,Palo Alto,LARGE,CONFIGURING,2023-02-01,24
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMVEs(testMVEs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMVEs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		mves        []*megaport.MVE
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			mves:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   location_id   location_name   vendor   size   provisioning_status   create_date   contract_term_months\n",
		},
		{
			name:        "empty slice",
			mves:        []*megaport.MVE{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
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
			expected:    "invalid MVE: nil value",
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
			expected:    "uid,name,location_id,location_name,vendor,size,provisioning_status,create_date,contract_term_months\n,,0,,,,,,0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printMVEs(tt.mves, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				switch tt.format {
				case "json":
					assert.JSONEq(t, tt.expected, output)
				case "table", "csv":
					assert.Equal(t, tt.expected, output)
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
				assert.Empty(t, output.LocationName)
				assert.Empty(t, output.Vendor)
				assert.Empty(t, output.Size)
				assert.Empty(t, output.ProvisioningStatus)
				assert.Empty(t, output.CreateDate)
				assert.Zero(t, output.ContractTermMonths)
			},
		},
		{
			name: "with location details",
			mve: &megaport.MVE{
				LocationDetails: &megaport.ProductLocationDetails{
					Name: "Test Location",
				},
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "Test Location", output.LocationName)
			},
		},
		{
			name: "with create date",
			mve: &megaport.MVE{
				CreateDate: &megaport.Time{Time: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC)},
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "2023-03-15", output.CreateDate)
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

// TestFilterMVEs tests the filterMVEs function directly
func TestFilterMVEs(t *testing.T) {
	mves := []*megaport.MVE{
		{
			UID:                "mve-123",
			Name:               "Production MVE",
			LocationID:         123,
			Vendor:             "Cisco",
			Size:               "MEDIUM",
			ProvisioningStatus: "LIVE",
		},
		{
			UID:                "mve-456",
			Name:               "Dev MVE",
			LocationID:         456,
			Vendor:             "Palo Alto",
			Size:               "LARGE",
			ProvisioningStatus: "LIVE",
		},
		{
			UID:                "mve-789",
			Name:               "Test Production",
			LocationID:         789,
			Vendor:             "Fortinet",
			Size:               "SMALL",
			ProvisioningStatus: "LIVE",
		},
	}

	tests := []struct {
		name         string
		nameFilter   string
		locationID   int
		vendorFilter string
		expectedIDs  []string
	}{
		{
			name:         "no filters",
			nameFilter:   "",
			locationID:   0,
			vendorFilter: "",
			expectedIDs:  []string{"mve-123", "mve-456", "mve-789"},
		},
		{
			name:         "filter by name - exact match",
			nameFilter:   "Dev MVE",
			locationID:   0,
			vendorFilter: "",
			expectedIDs:  []string{"mve-456"},
		},
		{
			name:         "filter by name - case insensitive substring match",
			nameFilter:   "production",
			locationID:   0,
			vendorFilter: "",
			expectedIDs:  []string{"mve-123", "mve-789"},
		},
		{
			name:         "filter by location ID",
			nameFilter:   "",
			locationID:   123,
			vendorFilter: "",
			expectedIDs:  []string{"mve-123"},
		},
		{
			name:         "filter by vendor",
			nameFilter:   "",
			locationID:   0,
			vendorFilter: "Palo Alto",
			expectedIDs:  []string{"mve-456"},
		},
		{
			name:         "filter by vendor - case insensitive",
			nameFilter:   "",
			locationID:   0,
			vendorFilter: "cisco",
			expectedIDs:  []string{"mve-123"},
		},
		{
			name:         "combine name and location filters",
			nameFilter:   "Production",
			locationID:   123,
			vendorFilter: "",
			expectedIDs:  []string{"mve-123"},
		},
		{
			name:         "combine name and vendor filters",
			nameFilter:   "Production",
			locationID:   0,
			vendorFilter: "Fortinet",
			expectedIDs:  []string{"mve-789"},
		},
		{
			name:         "combine all filters - no match",
			nameFilter:   "Production",
			locationID:   456,
			vendorFilter: "Cisco",
			expectedIDs:  []string{},
		},
		{
			name:         "no match with name filter",
			nameFilter:   "NonExistent",
			locationID:   0,
			vendorFilter: "",
			expectedIDs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterMVEs(mves, tt.nameFilter, tt.locationID, tt.vendorFilter)

			// Check count
			assert.Equal(t, len(tt.expectedIDs), len(filtered))

			// Check each expected ID is present
			filteredIDs := make([]string, len(filtered))
			for i, mve := range filtered {
				filteredIDs[i] = mve.UID
			}

			for _, expectedID := range tt.expectedIDs {
				assert.Contains(t, filteredIDs, expectedID)
			}
		})
	}
}
