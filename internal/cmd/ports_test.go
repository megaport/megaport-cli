package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testPorts = []*megaport.Port{
	{
		UID:                "port-1",
		Name:               "MyPortOne",
		LocationID:         1,
		PortSpeed:          1000,
		ProvisioningStatus: "ACTIVE",
	},
	{
		UID:                "port-2",
		Name:               "AnotherPort",
		LocationID:         2,
		PortSpeed:          2000,
		ProvisioningStatus: "INACTIVE",
	},
}

func TestFilterPorts(t *testing.T) {
	tests := []struct {
		name       string
		locationID int
		portSpeed  int
		portName   string
		expected   int
	}{
		{
			name:       "No filters",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			expected:   2,
		},
		{
			name:       "Filter by LocationID",
			locationID: 1,
			portSpeed:  0,
			portName:   "",
			expected:   1,
		},
		{
			name:       "Filter by PortSpeed",
			locationID: 0,
			portSpeed:  2000,
			portName:   "",
			expected:   1,
		},
		{
			name:       "Filter by PortName",
			locationID: 0,
			portSpeed:  0,
			portName:   "MyPortOne",
			expected:   1,
		},
		{
			name:       "No match",
			locationID: 99,
			portSpeed:  9999,
			portName:   "NoMatch",
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPorts(testPorts, tt.locationID, tt.portSpeed, tt.portName)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestPrintPorts_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "table")
		assert.NoError(t, err)
	})

	expected := `uid      name          location_id   port_speed   provisioning_status
port-1   MyPortOne     1             1000         ACTIVE
port-2   AnotherPort   2             2000         INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "port-1",
    "name": "MyPortOne",
    "location_id": 1,
    "port_speed": 1000,
    "provisioning_status": "ACTIVE"
  },
  {
    "uid": "port-2",
    "name": "AnotherPort",
    "location_id": 2,
    "port_speed": 2000,
    "provisioning_status": "INACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintPorts_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printPorts(testPorts, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
