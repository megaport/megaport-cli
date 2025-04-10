package partners

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var noColor = true // Disable color for testing

var testPartners = []*megaport.PartnerMegaport{
	{
		ProductUID:    "uid1",
		ProductName:   "ProductOne",
		ConnectType:   "TypeA",
		CompanyName:   "CompanyA",
		LocationId:    1,
		DiversityZone: "ZoneA",
		VXCPermitted:  true,
	},
	{
		ProductUID:    "uid2",
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
	output := output.CaptureOutput(func() {
		err := printPartnersFunc(testPartners, "table", noColor)
		assert.NoError(t, err)
	})

	expected := ` NAME       │ UID  │ CONNECT TYPE │ COMPANY NAME │ LOCATION ID │ DIVERSITY ZONE │ VXC PERMITTED 
────────────┼──────┼──────────────┼──────────────┼─────────────┼────────────────┼───────────────
 ProductOne │ uid1 │ TypeA        │ CompanyA     │ 1           │ ZoneA          │ true          
 ProductTwo │ uid2 │ TypeB        │ CompanyB     │ 2           │ ZoneB          │ false         
`
	assert.Equal(t, expected, output)
}

func TestPrintPartners_EmptySlice(t *testing.T) {
	var emptySlice []*megaport.PartnerMegaport

	// Table format with empty slice
	tableOutput := output.CaptureOutput(func() {
		err := printPartnersFunc(emptySlice, "table", noColor)
		assert.NoError(t, err)
	})
	expectedTable := ` NAME │ UID │ CONNECT TYPE │ COMPANY NAME │ LOCATION ID │ DIVERSITY ZONE │ VXC PERMITTED 
──────┼─────┼──────────────┼──────────────┼─────────────┼────────────────┼───────────────
`
	assert.Equal(t, expectedTable, tableOutput)

	// The rest of this function can remain unchanged
	// JSON format with empty slice
	jsonOutput := output.CaptureOutput(func() {
		err := printPartnersFunc(emptySlice, "json", noColor)
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", jsonOutput)

	// CSV format with empty slice
	csvOutput := output.CaptureOutput(func() {
		err := printPartnersFunc(emptySlice, "csv", noColor)
		assert.NoError(t, err)
	})
	expectedCSV := `product_name,uid,connect_type,company_name,location_id,diversity_zone,vxc_permitted
`
	assert.Equal(t, expectedCSV, csvOutput)
}

func TestPrintPartners_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printPartnersFunc(testPartners, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "product_name": "ProductOne",
    "uid": "uid1",
    "connect_type": "TypeA",
    "company_name": "CompanyA",
    "location_id": 1,
    "diversity_zone": "ZoneA",
    "vxc_permitted": true
  },
  {
    "product_name": "ProductTwo",
    "uid": "uid2",
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
	output := output.CaptureOutput(func() {
		err := printPartnersFunc(testPartners, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `product_name,uid,connect_type,company_name,location_id,diversity_zone,vxc_permitted
ProductOne,uid1,TypeA,CompanyA,1,ZoneA,true
ProductTwo,uid2,TypeB,CompanyB,2,ZoneB,false
`
	assert.Equal(t, expected, output)
}

func TestPrintPartners_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printPartnersFunc(testPartners, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
