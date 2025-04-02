package cmd

import (
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestPrintMCRs_Table(t *testing.T) {
	testMCRs := []*megaport.MCR{
		{
			UID:                "mcr-1",
			Name:               "MyMCROne",
			LocationID:         1,
			PortSpeed:          1000,
			ProvisioningStatus: "ACTIVE",
			ContractTermMonths: 12,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Sydney",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			UID:                "mcr-2",
			Name:               "AnotherMCR",
			LocationID:         2,
			PortSpeed:          5000,
			ProvisioningStatus: "INACTIVE",
			ContractTermMonths: 24,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Melbourne",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	output := captureOutput(func() {
		err := printMCRs(testMCRs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id   location_name   port_speed   provisioning_status   create_date   contract_term_months
mcr-1   MyMCROne     1             Sydney          1000         ACTIVE                2023-01-01    12
mcr-2   AnotherMCR   2             Melbourne       5000         INACTIVE              2023-02-01    24
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	testMCRs := []*megaport.MCR{
		{
			UID:                "mcr-1",
			Name:               "MyMCROne",
			LocationID:         1,
			PortSpeed:          1000,
			ProvisioningStatus: "ACTIVE",
			ContractTermMonths: 12,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Sydney",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			UID:                "mcr-2",
			Name:               "AnotherMCR",
			LocationID:         2,
			PortSpeed:          5000,
			ProvisioningStatus: "INACTIVE",
			ContractTermMonths: 24,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Melbourne",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	output := captureOutput(func() {
		err := printMCRs(testMCRs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mcr-1",
    "name": "MyMCROne",
    "location_id": 1,
    "location_name": "Sydney",
    "port_speed": 1000,
    "provisioning_status": "ACTIVE",
    "create_date": "2023-01-01",
    "contract_term_months": 12
  },
  {
    "uid": "mcr-2",
    "name": "AnotherMCR",
    "location_id": 2,
    "location_name": "Melbourne",
    "port_speed": 5000,
    "provisioning_status": "INACTIVE",
    "create_date": "2023-02-01",
    "contract_term_months": 24
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMCRs_CSV(t *testing.T) {
	// Updated test MCRs with location details
	testMCRs := []*megaport.MCR{
		{
			UID:                "mcr-1",
			Name:               "MyMCROne",
			LocationID:         1,
			PortSpeed:          1000,
			ProvisioningStatus: "ACTIVE",
			ContractTermMonths: 12,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Sydney",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			UID:                "mcr-2",
			Name:               "AnotherMCR",
			LocationID:         2,
			PortSpeed:          5000,
			ProvisioningStatus: "INACTIVE",
			ContractTermMonths: 24,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Melbourne",
			},
			CreateDate: &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	output := captureOutput(func() {
		err := printMCRs(testMCRs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,location_name,port_speed,provisioning_status,create_date,contract_term_months
mcr-1,MyMCROne,1,Sydney,1000,ACTIVE,2023-01-01,12
mcr-2,AnotherMCR,2,Melbourne,5000,INACTIVE,2023-02-01,24
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_EmptyAndNilSlice(t *testing.T) {
	tests := []struct {
		name     string
		mcrs     []*megaport.MCR
		format   string
		expected string
	}{
		{
			name:   "empty slice table format",
			mcrs:   []*megaport.MCR{},
			format: "table",
			expected: `uid   name   location_id   location_name   port_speed   provisioning_status   create_date   contract_term_months
`,
		},
		{
			name:   "empty slice csv format",
			mcrs:   []*megaport.MCR{},
			format: "csv",
			expected: `uid,name,location_id,location_name,port_speed,provisioning_status,create_date,contract_term_months
`,
		},
		{
			name:     "empty slice json format",
			mcrs:     []*megaport.MCR{},
			format:   "json",
			expected: "[]\n",
		},
		{
			name:   "nil slice table format",
			mcrs:   nil,
			format: "table",
			expected: `uid   name   location_id   location_name   port_speed   provisioning_status   create_date   contract_term_months
`,
		},
		{
			name:   "nil slice csv format",
			mcrs:   nil,
			format: "csv",
			expected: `uid,name,location_id,location_name,port_speed,provisioning_status,create_date,contract_term_months
`,
		},
		{
			name:     "nil slice json format",
			mcrs:     nil,
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(tt.mcrs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

// TestFilterMCRs tests the filterMCRs function directly
func TestFilterMCRs(t *testing.T) {
	mcrs := []*megaport.MCR{
		{
			UID:                "mcr-123",
			Name:               "Production MCR",
			LocationID:         123,
			PortSpeed:          1000,
			ProvisioningStatus: "LIVE",
		},
		{
			UID:                "mcr-456",
			Name:               "Dev MCR",
			LocationID:         456,
			PortSpeed:          5000,
			ProvisioningStatus: "LIVE",
		},
		{
			UID:                "mcr-789",
			Name:               "Test Production",
			LocationID:         789,
			PortSpeed:          10000,
			ProvisioningStatus: "LIVE",
		},
	}

	tests := []struct {
		name        string
		nameFilter  string
		locationID  int
		portSpeed   int
		expectedIDs []string
	}{
		{
			name:        "no filters",
			nameFilter:  "",
			locationID:  0,
			portSpeed:   0,
			expectedIDs: []string{"mcr-123", "mcr-456", "mcr-789"},
		},
		{
			name:        "filter by name - exact match",
			nameFilter:  "Dev MCR",
			locationID:  0,
			portSpeed:   0,
			expectedIDs: []string{"mcr-456"},
		},
		{
			name:        "filter by name - case insensitive substring match",
			nameFilter:  "production",
			locationID:  0,
			portSpeed:   0,
			expectedIDs: []string{"mcr-123", "mcr-789"},
		},
		{
			name:        "filter by location ID",
			nameFilter:  "",
			locationID:  123,
			portSpeed:   0,
			expectedIDs: []string{"mcr-123"},
		},
		{
			name:        "filter by port speed",
			nameFilter:  "",
			locationID:  0,
			portSpeed:   5000,
			expectedIDs: []string{"mcr-456"},
		},
		{
			name:        "combine name and location filters",
			nameFilter:  "Production",
			locationID:  123,
			portSpeed:   0,
			expectedIDs: []string{"mcr-123"},
		},
		{
			name:        "combine name and port speed filters",
			nameFilter:  "Production",
			locationID:  0,
			portSpeed:   10000,
			expectedIDs: []string{"mcr-789"},
		},
		{
			name:        "combine all filters - no match",
			nameFilter:  "Production",
			locationID:  456,
			portSpeed:   10000,
			expectedIDs: []string{},
		},
		{
			name:        "no match with name filter",
			nameFilter:  "NonExistent",
			locationID:  0,
			portSpeed:   0,
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterMCRs(mcrs, tt.nameFilter, tt.locationID, tt.portSpeed)

			// Check count
			assert.Equal(t, len(tt.expectedIDs), len(filtered))

			// Check each expected ID is present
			filteredIDs := make([]string, len(filtered))
			for i, mcr := range filtered {
				filteredIDs[i] = mcr.UID
			}

			for _, expectedID := range tt.expectedIDs {
				assert.Contains(t, filteredIDs, expectedID)
			}
		})
	}
}
