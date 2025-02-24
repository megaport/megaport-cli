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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPartners(testPartners, tt.productName, tt.connectType, tt.companyName, tt.locationID, tt.diversityZone)
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

func TestPrintPartners_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printPartners(testPartners, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
