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
		printPartners(testPartners, "table")
	})

	// Table output should contain headers and both partner product names
	assert.Contains(t, output, "ProductName")
	assert.Contains(t, output, "ProductOne")
	assert.Contains(t, output, "ProductTwo")
	assert.Contains(t, output, "TypeA")
	assert.Contains(t, output, "TypeB")
	assert.Contains(t, output, "CompanyA")
	assert.Contains(t, output, "CompanyB")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "ZoneA")
	assert.Contains(t, output, "ZoneB")
	assert.Contains(t, output, "true")
	assert.Contains(t, output, "false")
}

func TestPrintPartners_JSON(t *testing.T) {
	output := captureOutput(func() {
		printPartners(testPartners, "json")
	})

	// JSON output should contain an array of objects
	assert.Contains(t, output, `"product_name":"ProductOne"`)
	assert.Contains(t, output, `"product_name":"ProductTwo"`)
	assert.Contains(t, output, `"connect_type":"TypeA"`)
	assert.Contains(t, output, `"connect_type":"TypeB"`)
	assert.Contains(t, output, `"company_name":"CompanyA"`)
	assert.Contains(t, output, `"company_name":"CompanyB"`)
	assert.Contains(t, output, `"location_id":1`)
	assert.Contains(t, output, `"location_id":2`)
	assert.Contains(t, output, `"diversity_zone":"ZoneA"`)
	assert.Contains(t, output, `"diversity_zone":"ZoneB"`)
	assert.Contains(t, output, `"vxc_permitted":true`)
	assert.Contains(t, output, `"vxc_permitted":false`)
}

func TestPrintPartners_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printPartners(testPartners, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
