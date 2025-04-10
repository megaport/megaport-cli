package mcr

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var noColor = true // Disable color for testing

var testMCRs = []*megaport.MCR{
	{
		UID:                "mcr-1",
		Name:               "MyMCROne",
		LocationID:         1,
		ProvisioningStatus: "ACTIVE",
		PortSpeed:          1000,
		Resources: megaport.MCRResources{
			VirtualRouter: megaport.MCRVirtualRouter{
				ASN: 64512,
			},
		},
	},
	{
		UID:                "mcr-2",
		Name:               "AnotherMCR",
		LocationID:         2,
		ProvisioningStatus: "INACTIVE",
		PortSpeed:          5000,
		Resources: megaport.MCRResources{
			VirtualRouter: megaport.MCRVirtualRouter{
				ASN: 64513,
			},
		},
	},
}

func TestPrintMCRs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "table", noColor)
		assert.NoError(t, err)
	})

	expected := ` UID   │ NAME       │ LOCATION ID │ STATUS   │ ASN   │ SPEED 
───────┼────────────┼─────────────┼──────────┼───────┼───────
 mcr-1 │ MyMCROne   │ 1           │ ACTIVE   │ 64512 │ 1000  
 mcr-2 │ AnotherMCR │ 2           │ INACTIVE │ 64513 │ 5000  
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mcr-1",
    "name": "MyMCROne",
    "location_id": 1,
    "provisioning_status": "ACTIVE",
    "asn": 64512,
    "speed": 1000
  },
  {
    "uid": "mcr-2",
    "name": "AnotherMCR",
    "location_id": 2,
    "provisioning_status": "INACTIVE",
    "asn": 64513,
    "speed": 5000
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMCRs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,provisioning_status,asn,speed
mcr-1,MyMCROne,1,ACTIVE,64512,1000
mcr-2,AnotherMCR,2,INACTIVE,64513,5000
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printMCRs(testMCRs, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
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
			expected: ` UID │ NAME │ LOCATION ID │ STATUS │ ASN │ SPEED 
─────┼──────┼─────────────┼────────┼─────┼───────
`,
		},
		{
			name:   "empty slice csv format",
			mcrs:   []*megaport.MCR{},
			format: "csv",
			expected: `uid,name,location_id,provisioning_status,asn,speed
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
			expected: ` UID │ NAME │ LOCATION ID │ STATUS │ ASN │ SPEED 
─────┼──────┼─────────────┼────────┼─────┼───────
`,
		},
		{
			name:   "nil slice csv format",
			mcrs:   nil,
			format: "csv",
			expected: `uid,name,location_id,provisioning_status,asn,speed
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
			output := output.CaptureOutput(func() {
				err := printMCRs(tt.mcrs, tt.format, noColor)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}
