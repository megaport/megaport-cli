package cmd

import (
	"testing"

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
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id   provisioning_status
mcr-1   MyMCROne     1             ACTIVE
mcr-2   AnotherMCR   2             INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "json")
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
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,provisioning_status
mcr-1,MyMCROne,1,ACTIVE
mcr-2,AnotherMCR,2,INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMCRs(testMCRs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMCRs_EmptySlice(t *testing.T) {
	var emptyMCRs []*megaport.MCR

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "table format",
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "csv format",
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "json format",
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(emptyMCRs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestPrintMCRs_NilSlice(t *testing.T) {
	var nilMCRs []*megaport.MCR = nil

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "table format",
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "csv format",
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "json format",
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(nilMCRs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestPrintMCRs_InvalidMCR(t *testing.T) {
	invalidMCRs := []*megaport.MCR{
		{
			UID:                "",
			Name:               "",
			LocationID:         0,
			ProvisioningStatus: "",
		},
	}

	tests := []struct {
		name        string
		format      string
		mcrs        []*megaport.MCR
		shouldError bool
		expected    string
	}{
		{
			name:        "table format with zero values",
			format:      "table",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    "   0             ",
		},
		{
			name:        "csv format with zero values",
			format:      "csv",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    ",,0,",
		},
		{
			name:        "json format with zero values",
			format:      "json",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    `[{"uid":"","name":"","location_id":0,"provisioning_status":""}]`,
		},
		{
			name:        "table format with nil MCR",
			format:      "table",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
		{
			name:        "csv format with nil MCR",
			format:      "csv",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
		{
			name:        "json format with nil MCR",
			format:      "json",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printMCRs(tt.mcrs, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				if tt.format == "json" {
					assert.JSONEq(t, tt.expected, output)
				} else {
					assert.Contains(t, output, tt.expected)
				}
			}
		})
	}
}
