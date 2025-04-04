package cmd

import (
	"encoding/json"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testPorts = []*megaport.Port{
	{
		UID:                "port-1",
		Name:               "MyPortOne",
		LocationID:         1,
		LocationDetails:    &megaport.ProductLocationDetails{Name: "Sydney"},
		PortSpeed:          1000,
		ProvisioningStatus: "ACTIVE",
		CreateDate:         &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		ContractTermMonths: 12,
	},
	{
		UID:                "port-2",
		Name:               "AnotherPort",
		LocationID:         2,
		LocationDetails:    &megaport.ProductLocationDetails{Name: "Melbourne"},
		PortSpeed:          2000,
		ProvisioningStatus: "INACTIVE",
		CreateDate:         &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		ContractTermMonths: 24,
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

	expected := `uid      name          location_id   location_name   port_speed   provisioning_status   create_date   contract_term_months
port-1   MyPortOne     1             Sydney          1000         ACTIVE                2023-01-01    12
port-2   AnotherPort   2             Melbourne       2000         INACTIVE              2023-02-01    24
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "json")
		assert.NoError(t, err)
	})

	var jsonOutput []map[string]interface{}
	err := json.Unmarshal([]byte(output), &jsonOutput)
	assert.NoError(t, err)

	expected := []map[string]interface{}{
		{
			"uid":                  "port-1",
			"name":                 "MyPortOne",
			"location_id":          float64(1),
			"location_name":        "Sydney",
			"port_speed":           float64(1000),
			"provisioning_status":  "ACTIVE",
			"create_date":          "2023-01-01",
			"contract_term_months": float64(12),
		},
		{
			"uid":                  "port-2",
			"name":                 "AnotherPort",
			"location_id":          float64(2),
			"location_name":        "Melbourne",
			"port_speed":           float64(2000),
			"provisioning_status":  "INACTIVE",
			"create_date":          "2023-02-01",
			"contract_term_months": float64(24),
		},
	}
	assert.Equal(t, expected, jsonOutput)
}

func TestPrintPorts_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,location_name,port_speed,provisioning_status,create_date,contract_term_months
port-1,MyPortOne,1,Sydney,1000,ACTIVE,2023-01-01,12
port-2,AnotherPort,2,Melbourne,2000,INACTIVE,2023-02-01,24
`
	assert.Equal(t, expected, output)
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
			expected:    "uid   name   location_id   location_name   port_speed   provisioning_status   create_date   contract_term_months\n",
		},
		{
			name:        "empty slice",
			ports:       []*megaport.Port{},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id,location_name,port_speed,provisioning_status,create_date,contract_term_months\n",
		},
		{
			name: "port with zero values",
			ports: []*megaport.Port{
				{},
			},
			format:      "json",
			shouldError: false,
			expected:    `[{"uid":"","name":"","location_id":0,"location_name":"","port_speed":0,"provisioning_status":"","create_date":"","contract_term_months":0}]`,
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

			output = captureOutput(func() {
				err = printPorts(tt.ports, tt.format)
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
