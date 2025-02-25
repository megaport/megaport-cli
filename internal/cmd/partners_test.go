package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testPartners = []*megaport.PartnerMegaport{
	{
		ProductName:   "ProductOne",
		ConnectType:   "TypeA",
		CompanyName:   "CompanyA",
		LocationId:    1,
		DiversityZone: "ZoneA",
		VXCPermitted:  true,
	},
	{
		ProductName:   "ProductTwo",
		ConnectType:   "TypeB",
		CompanyName:   "CompanyB",
		LocationId:    2,
		DiversityZone: "ZoneB",
		VXCPermitted:  false,
	},
}

func TestFilterPartners(t *testing.T) {
	tests := []struct {
		name          string
		productName   string
		connectType   string
		companyName   string
		locationID    int
		diversityZone string
		expected      int
	}{
		{
			name:          "No filters",
			productName:   "",
			connectType:   "",
			companyName:   "",
			locationID:    0,
			diversityZone: "",
			expected:      2,
		},
		{
			name:          "Filter by ProductName",
			productName:   "ProductOne",
			connectType:   "",
			companyName:   "",
			locationID:    0,
			diversityZone: "",
			expected:      1,
		},
		{
			name:          "Filter by ConnectType",
			productName:   "",
			connectType:   "TypeB",
			companyName:   "",
			locationID:    0,
			diversityZone: "",
			expected:      1,
		},
		{
			name:          "Filter by CompanyName",
			productName:   "",
			connectType:   "",
			companyName:   "CompanyA",
			locationID:    0,
			diversityZone: "",
			expected:      1,
		},
		{
			name:          "Filter by LocationID",
			productName:   "",
			connectType:   "",
			companyName:   "",
			locationID:    1,
			diversityZone: "",
			expected:      1,
		},
		{
			name:          "Filter by DiversityZone",
			productName:   "",
			connectType:   "",
			companyName:   "",
			locationID:    0,
			diversityZone: "ZoneB",
			expected:      1,
		},
		{
			name:          "No match",
			productName:   "NoMatch",
			connectType:   "NoMatch",
			companyName:   "NoMatch",
			locationID:    99,
			diversityZone: "NoMatch",
			expected:      0,
		},
		{
			name:          "Empty partners slice",
			productName:   "ProductOne",
			connectType:   "",
			companyName:   "",
			locationID:    0,
			diversityZone: "",
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var source []*megaport.PartnerMegaport
			if tt.name == "Empty partners slice" {
				source = []*megaport.PartnerMegaport{}
			} else {
				source = testPartners
			}

			result := filterPartners(
				source,
				tt.productName,
				tt.connectType,
				tt.companyName,
				tt.locationID,
				tt.diversityZone,
			)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestPrintPartners_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printPartners(testPartners, "table")
		assert.NoError(t, err)
	})

	expected := `product_name   connect_type   company_name   location_id   diversity_zone   vxc_permitted
ProductOne     TypeA          CompanyA       1             ZoneA            true
ProductTwo     TypeB          CompanyB       2             ZoneB            false
`
	assert.Equal(t, expected, output)
}

func TestPrintPartners_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printPartners(testPartners, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "product_name": "ProductOne",
    "connect_type": "TypeA",
    "company_name": "CompanyA",
    "location_id": 1,
    "diversity_zone": "ZoneA",
    "vxc_permitted": true
  },
  {
    "product_name": "ProductTwo",
    "connect_type": "TypeB",
    "company_name": "CompanyB",
    "location_id": 2,
    "diversity_zone": "ZoneB",
    "vxc_permitted": false
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintPartners_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printPartners(testPartners, "csv")
		assert.NoError(t, err)
	})

	expected := `product_name,connect_type,company_name,location_id,diversity_zone,vxc_permitted
ProductOne,TypeA,CompanyA,1,ZoneA,true
ProductTwo,TypeB,CompanyB,2,ZoneB,false
`
	assert.Equal(t, expected, output)
}

func TestPrintPartners_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printPartners(testPartners, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

// Additional coverage test: printing an empty slice
func TestPrintPartners_EmptySlice(t *testing.T) {
	var emptySlice []*megaport.PartnerMegaport

	// Table format with empty slice
	tableOutput := captureOutput(func() {
		err := printPartners(emptySlice, "table")
		assert.NoError(t, err)
	})
	// Should only print the header row
	expectedTable := `product_name   connect_type   company_name   location_id   diversity_zone   vxc_permitted
`
	assert.Equal(t, expectedTable, tableOutput)

	// JSON format with empty slice
	jsonOutput := captureOutput(func() {
		err := printPartners(emptySlice, "json")
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", jsonOutput)

	// CSV format with empty slice
	csvOutput := captureOutput(func() {
		err := printPartners(emptySlice, "csv")
		assert.NoError(t, err)
	})
	expectedCSV := `product_name,connect_type,company_name,location_id,diversity_zone,vxc_permitted
`
	assert.Equal(t, expectedCSV, csvOutput)
}
