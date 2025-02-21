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
		printPorts(testPorts, "table")
	})

	// Table output should contain headers and both port names
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "LocationID")
	assert.Contains(t, output, "PortSpeed")
	assert.Contains(t, output, "ProvisioningStatus")
	assert.Contains(t, output, "MyPortOne")
	assert.Contains(t, output, "AnotherPort")
}

func TestPrintPorts_JSON(t *testing.T) {
	output := captureOutput(func() {
		printPorts(testPorts, "json")
	})

	// JSON output should contain an array of objects with the correct keys
	assert.Contains(t, output, `"uid":"port-1"`)
	assert.Contains(t, output, `"uid":"port-2"`)
	assert.Contains(t, output, `"name":"MyPortOne"`)
	assert.Contains(t, output, `"name":"AnotherPort"`)
	assert.Contains(t, output, `"location_id":1`)
	assert.Contains(t, output, `"location_id":2`)
	assert.Contains(t, output, `"port_speed":1000`)
	assert.Contains(t, output, `"port_speed":2000`)
	assert.Contains(t, output, `"provisioning_status":"ACTIVE"`)
	assert.Contains(t, output, `"provisioning_status":"INACTIVE"`)
}

func TestPrintPorts_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printPorts(testPorts, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
