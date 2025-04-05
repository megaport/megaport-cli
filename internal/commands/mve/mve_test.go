package mve

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:        "mve-1",
		Name:       "MyMVEOne",
		LocationID: 1,
	},
	{
		UID:        "mve-2",
		Name:       "AnotherMVE",
		LocationID: 2,
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "table", noColor)
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id
mve-1   MyMVEOne     1
mve-2   AnotherMVE   2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printMVEs(testMVEs, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id
mve-1,MyMVEOne,1
mve-2,AnotherMVE,2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printMVEs(testMVEs, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMVEs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		mves        []*megaport.MVE
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			mves:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   location_id\n",
		},
		{
			name:        "empty slice",
			mves:        []*megaport.MVE{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil mve in slice",
			mves: []*megaport.MVE{
				nil,
				{
					UID:        "mve-1",
					Name:       "TestMVE",
					LocationID: 1,
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mves: []*megaport.MVE{
				{
					UID:        "",
					Name:       "",
					LocationID: 0,
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id\n,,0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var op string
			var err error

			op = output.CaptureOutput(func() {
				err = printMVEs(tt.mves, tt.format, noColor)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, op)
			} else {
				assert.NoError(t, err)
				switch tt.format {
				case "json":
					assert.JSONEq(t, tt.expected, op)
				case "table", "csv":
					assert.Equal(t, tt.expected, op)
				}
			}
		})
	}
}

func TestToMVEOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		mve           *megaport.MVE
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, MVEOutput)
	}{
		{
			name:          "nil mve",
			mve:           nil,
			shouldError:   true,
			errorContains: "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mve:  &megaport.MVE{},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
		{
			name: "whitespace values",
			mve: &megaport.MVE{
				UID:        "   ",
				Name:       "   ",
				LocationID: 0,
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "   ", output.UID)
				assert.Equal(t, "   ", output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToMVEOutput(tt.mve)

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
