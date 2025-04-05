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
	},
}

func TestPrintVXCs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "table", true)
		assert.NoError(t, err)
	})

	expected := `uid     name         a_end_uid   b_end_uid
vxc-1   MyVXCOne     a-end-1     b-end-1
vxc-2   AnotherVXC   a-end-2     b-end-2
`
	assert.Equal(t, expected, output)
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
    "b_end_uid": "b-end-1"
  },
  {
    "uid": "vxc-2",
    "name": "AnotherVXC",
    "a_end_uid": "a-end-2",
    "b_end_uid": "b-end-2"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintVXCs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "csv", true)
		assert.NoError(t, err)
	})

	expected := `uid,name,a_end_uid,b_end_uid
vxc-1,MyVXCOne,a-end-1,b-end-1
vxc-2,AnotherVXC,a-end-2,b-end-2
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
		name        string
		vxcs        []*megaport.VXC
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			vxcs:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   a_end_uid   b_end_uid\n",
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
					UID:  "vxc-1",
					Name: "TestVXC",
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,a_end_uid,b_end_uid\nvxc-1,TestVXC,,\n",
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
				switch tt.format {
				case "json":
					assert.JSONEq(t, tt.expected, capturedOutput)
				case "table", "csv":
					assert.Equal(t, tt.expected, capturedOutput)
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
			},
		},
		{
			name: "nil end configurations",
			vxc: &megaport.VXC{
				UID:  "vxc-1",
				Name: "TestVXC",
			},
			validateFunc: func(t *testing.T, output VXCOutput) {
				assert.Equal(t, "vxc-1", output.UID)
				assert.Equal(t, "TestVXC", output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
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
