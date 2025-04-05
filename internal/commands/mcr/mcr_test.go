package mcr

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMCRs = []*megaport.MCR{
	{
		UID:                "mcr-1",
		Name:               "MyMCROne",
		LocationID:         1,
		ProvisioningStatus: "ACTIVE",
	},
	{
		UID:                "mcr-2",
		Name:               "AnotherMCR",
		LocationID:         2,
		ProvisioningStatus: "INACTIVE",
	},
}

func TestPrintMCRs_Table(t *testing.T) {
	noColor := true
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "table", noColor)
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id   provisioning_status
mcr-1   MyMCROne     1             ACTIVE
mcr-2   AnotherMCR   2             INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	noColor := true
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mcr-1",
    "name": "MyMCROne",
    "location_id": 1,
    "provisioning_status": "ACTIVE"
  },
  {
    "uid": "mcr-2",
    "name": "AnotherMCR",
    "location_id": 2,
    "provisioning_status": "INACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMCRs_CSV(t *testing.T) {
	noColor := true
	output := output.CaptureOutput(func() {
		err := printMCRs(testMCRs, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,provisioning_status
mcr-1,MyMCROne,1,ACTIVE
mcr-2,AnotherMCR,2,INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	noColor := true
	var err error
	output := output.CaptureOutput(func() {
		err = printMCRs(testMCRs, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMCRs_EmptyAndNilSlice(t *testing.T) {
	noColor := true
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
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "empty slice csv format",
			mcrs:   []*megaport.MCR{},
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
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
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "nil slice csv format",
			mcrs:   nil,
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
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
