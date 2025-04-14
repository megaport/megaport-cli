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

	// Check for headers and content
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "CONNECT TYPE")
	assert.Contains(t, output, "COMPANY NAME")
	assert.Contains(t, output, "LOCATION ID")
	assert.Contains(t, output, "DIVERSITY ZONE")
	assert.Contains(t, output, "VXC PERMITTED")

	// Check for actual data
	assert.Contains(t, output, "ProductOne")
	assert.Contains(t, output, "uid1")
	assert.Contains(t, output, "TypeA")
	assert.Contains(t, output, "CompanyA")
	assert.Contains(t, output, "ZoneA")
	assert.Contains(t, output, "true")

	assert.Contains(t, output, "ProductTwo")
	assert.Contains(t, output, "uid2")
	assert.Contains(t, output, "TypeB")
	assert.Contains(t, output, "CompanyB")
	assert.Contains(t, output, "ZoneB")
	assert.Contains(t, output, "false")

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

func TestPrintPartners_EmptySlice(t *testing.T) {
	var emptySlice []*megaport.PartnerMegaport

	// Table format with empty slice
	tableOutput := output.CaptureOutput(func() {
		err := printPartnersFunc(emptySlice, "table", noColor)
		assert.NoError(t, err)
	})

	// Check for headers and box drawing characters in empty table
	assert.Contains(t, tableOutput, "NAME")
	assert.Contains(t, tableOutput, "UID")
	assert.Contains(t, tableOutput, "CONNECT TYPE")
	assert.Contains(t, tableOutput, "COMPANY NAME")
	assert.Contains(t, tableOutput, "LOCATION ID")
	assert.Contains(t, tableOutput, "DIVERSITY ZONE")
	assert.Contains(t, tableOutput, "VXC PERMITTED")

	// Check for box drawing characters
	assert.Contains(t, tableOutput, "┌")
	assert.Contains(t, tableOutput, "┐")
	assert.Contains(t, tableOutput, "└")
	assert.Contains(t, tableOutput, "┘")
	assert.Contains(t, tableOutput, "│")
	assert.Contains(t, tableOutput, "─")

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
