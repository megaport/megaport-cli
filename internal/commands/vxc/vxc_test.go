package vxc

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testVXCs = []*megaport.VXC{
	{
		UID:  "vxc-1",
		Name: "MyVXCOne",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-1",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-1",
		},
		RateLimit:          100,
		ProvisioningStatus: "CONFIGURED",
	},
	{
		UID:  "vxc-2",
		Name: "AnotherVXC",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-2",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-2",
		},
		RateLimit:          200,
		ProvisioningStatus: "LIVE",
	},
}

func TestPrintVXCs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "table", true)
		assert.NoError(t, err)
	})

	// Check for headers and content
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "A END UID")
	assert.Contains(t, output, "B END UID")
	assert.Contains(t, output, "A END VLAN")
	assert.Contains(t, output, "B END VLAN")
	assert.Contains(t, output, "RATE LIMIT")
	assert.Contains(t, output, "STATUS")

	// Check for actual data
	assert.Contains(t, output, "vxc-1")
	assert.Contains(t, output, "MyVXCOne")
	assert.Contains(t, output, "a-end-1")
	assert.Contains(t, output, "b-end-1")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "CONFIGURED")

	assert.Contains(t, output, "vxc-2")
	assert.Contains(t, output, "AnotherVXC")
	assert.Contains(t, output, "a-end-2")
	assert.Contains(t, output, "b-end-2")
	assert.Contains(t, output, "200")
	assert.Contains(t, output, "LIVE")

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

func TestPrintVXCs_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "json", true)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "vxc-1",
    "name": "MyVXCOne",
    "a_end_uid": "a-end-1",
    "b_end_uid": "b-end-1",
    "a_end_vlan": 0,
    "b_end_vlan": 0,
    "rate_limit": 100,
    "status": "CONFIGURED"
  },
  {
    "uid": "vxc-2",
    "name": "AnotherVXC",
    "a_end_uid": "a-end-2",
    "b_end_uid": "b-end-2",
    "a_end_vlan": 0,
    "b_end_vlan": 0,
    "rate_limit": 200,
    "status": "LIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintVXCs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "csv", true)
		assert.NoError(t, err)
	})

	expected := `uid,name,a_end_uid,b_end_uid,a_end_vlan,b_end_vlan,rate_limit,status
vxc-1,MyVXCOne,a-end-1,b-end-1,0,0,100,CONFIGURED
vxc-2,AnotherVXC,a-end-2,b-end-2,0,0,200,LIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintVXCs_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printVXCs(testVXCs, "invalid", true)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintVXCs_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		vxcs         []*megaport.VXC
		format       string
		shouldError  bool
		validateFunc func(*testing.T, string) // New function to validate output
		expected     string                   // Keep for JSON and CSV validation
	}{
		{
			name:        "nil slice",
			vxcs:        nil,
			format:      "table",
			shouldError: false,
			validateFunc: func(t *testing.T, output string) {
				// Check for headers and box drawing characters in empty table
				assert.Contains(t, output, "UID")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "A END UID")
				assert.Contains(t, output, "B END UID")
				assert.Contains(t, output, "A END VLAN")
				assert.Contains(t, output, "B END VLAN")
				assert.Contains(t, output, "RATE LIMIT")
				assert.Contains(t, output, "STATUS")

				// Check for box drawing characters
				assert.Contains(t, output, "┌")
				assert.Contains(t, output, "┐")
				assert.Contains(t, output, "└")
				assert.Contains(t, output, "┘")
				assert.Contains(t, output, "│")
				assert.Contains(t, output, "─")
			},
		},
		{
			name:        "empty slice",
			vxcs:        []*megaport.VXC{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil vxc in slice",
			vxcs: []*megaport.VXC{
				nil,
				{
					UID:  "vxc-1",
					Name: "TestVXC",
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid VXC: nil value",
		},
		{
			name: "nil end configurations",
			vxcs: []*megaport.VXC{
				{
					UID:                "vxc-1",
					Name:               "TestVXC",
					RateLimit:          50,
					ProvisioningStatus: "PENDING",
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,a_end_uid,b_end_uid,a_end_vlan,b_end_vlan,rate_limit,status\nvxc-1,TestVXC,,,0,0,50,PENDING\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			capturedOutput := output.CaptureOutput(func() {
				err = printVXCs(tt.vxcs, tt.format, true)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, capturedOutput)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, capturedOutput)
				} else if tt.expected != "" {
					switch tt.format {
					case "json":
						assert.JSONEq(t, tt.expected, capturedOutput)
					case "csv":
						assert.Equal(t, tt.expected, capturedOutput)
					}
				}
			}
		})
	}
}

func TestToVXCOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		vxc           *megaport.VXC
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, VXCOutput)
	}{
		{
			name:          "nil vxc",
			vxc:           nil,
			shouldError:   true,
			errorContains: "invalid VXC: nil value",
		},
		{
			name: "zero values",
			vxc:  &megaport.VXC{},
			validateFunc: func(t *testing.T, output VXCOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
				assert.Equal(t, 0, output.AEndVLAN)
				assert.Equal(t, 0, output.BEndVLAN)
				assert.Equal(t, 0, output.RateLimit)
				assert.Empty(t, output.Status)
			},
		},
		{
			name: "nil end configurations",
			vxc: &megaport.VXC{
				UID:                "vxc-1",
				Name:               "TestVXC",
				RateLimit:          75,
				ProvisioningStatus: "CONFIGURED",
			},
			validateFunc: func(t *testing.T, output VXCOutput) {
				assert.Equal(t, "vxc-1", output.UID)
				assert.Equal(t, "TestVXC", output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
				assert.Equal(t, 0, output.AEndVLAN)
				assert.Equal(t, 0, output.BEndVLAN)
				assert.Equal(t, 75, output.RateLimit)
				assert.Equal(t, "CONFIGURED", output.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToVXCOutput(tt.vxc)

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
