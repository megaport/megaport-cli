package ports

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
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
	noColor := false
	output := op.CaptureOutput(func() {
		err := printPorts(testPorts, "table", noColor)
		assert.NoError(t, err)
	})

	expected := `UID      Name          LocationID   Speed   Status
port-1   MyPortOne     1            1000    ACTIVE
port-2   AnotherPort   2            2000    INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_JSON(t *testing.T) {
	noColor := true
	output := op.CaptureOutput(func() {
		err := printPorts(testPorts, "json", noColor)
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

func TestPrintPorts_CSV(t *testing.T) {
	noColor := true
	output := op.CaptureOutput(func() {
		err := printPorts(testPorts, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,port_speed,provisioning_status
port-1,MyPortOne,1,1000,ACTIVE
port-2,AnotherPort,2,2000,INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_Invalid(t *testing.T) {
	var err error
	noColor := false
	output := op.CaptureOutput(func() {
		err = printPorts(testPorts, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintPorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		ports       []*megaport.Port
		format      string
		shouldError bool
		expected    string
		contains    string // New field for partial matches
	}{
		{
			name:        "nil slice",
			ports:       nil,
			format:      "table",
			shouldError: false,
			expected:    "UID   Name   LocationID   Speed   Status\n",
		},
		{
			name:        "empty slice",
			ports:       []*megaport.Port{},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id,port_speed,provisioning_status\n",
		},
		{
			name: "port with zero values",
			ports: []*megaport.Port{
				{
					UID:                "",
					Name:               "",
					LocationID:         0,
					PortSpeed:          0,
					ProvisioningStatus: "",
				},
			},
			format:      "json",
			shouldError: false,
			expected:    `[{"uid":"","name":"","location_id":0,"port_speed":0,"provisioning_status":""}]`,
		},
		{
			name:        "nil port in slice",
			ports:       []*megaport.Port{nil},
			format:      "table",
			shouldError: true,
			expected:    "invalid port: nil value",
		},
		{
			name: "mixed valid and nil ports",
			ports: []*megaport.Port{
				{
					UID:                "port-1",
					Name:               "ValidPort",
					LocationID:         1,
					PortSpeed:          1000,
					ProvisioningStatus: "ACTIVE",
				},
				nil,
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid port: nil value",
		},
		{
			name: "port with invalid status",
			ports: []*megaport.Port{
				{
					UID:                "port-1",
					Name:               "TestPort",
					LocationID:         1,
					PortSpeed:          1000,
					ProvisioningStatus: "INVALID_STATUS",
				},
			},
			format:      "table",
			shouldError: false,
			contains:    "INVALID_STATUS", // We just want to check if this status appears
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error
			var noColor bool

			output = op.CaptureOutput(func() {
				err = printPorts(tt.ports, tt.format, noColor)
			})

			if tt.shouldError {
				assert.Error(t, err)
				if tt.expected != "" {
					assert.Contains(t, err.Error(), tt.expected)
				}
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					if tt.format == "json" {
						assert.JSONEq(t, tt.expected, output)
					} else {
						assert.Equal(t, tt.expected, output)
					}
				}
				if tt.contains != "" {
					assert.Contains(t, output, tt.contains)
				}
			}
		})
	}
}

func TestFilterPorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		ports      []*megaport.Port
		locationID int
		portSpeed  int
		portName   string
		expected   int
	}{
		{
			name:       "nil slice",
			ports:      nil,
			locationID: 1,
			portSpeed:  1000,
			portName:   "Test",
			expected:   0,
		},
		{
			name:       "empty slice",
			ports:      []*megaport.Port{},
			locationID: 1,
			portSpeed:  1000,
			portName:   "Test",
			expected:   0,
		},
		{
			name: "slice with nil port",
			ports: []*megaport.Port{
				nil,
				{
					UID:       "port-1",
					Name:      "TestPort",
					PortSpeed: 1000,
				},
			},
			locationID: 0,
			portSpeed:  1000,
			portName:   "",
			expected:   1, // Should skip nil and return valid port
		},
		{
			name: "zero values in port",
			ports: []*megaport.Port{
				{
					UID:       "",
					Name:      "",
					PortSpeed: 0,
				},
			},
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			expected:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPorts(tt.ports, tt.locationID, tt.portSpeed, tt.portName)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}
