package ports

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFlagPortInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "valid input",
			flags: map[string]string{
				"name":                   "test-port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "1",
				"marketplace-visibility": "true",
			},
		},
		{
			name: "with resource tags",
			flags: map[string]string{
				"name":                   "test-port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "1",
				"marketplace-visibility": "true",
				"resource-tags":          `{"env":"prod"}`,
			},
		},
		{
			name: "invalid resource tags JSON",
			flags: map[string]string{
				"name":                   "test-port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "1",
				"marketplace-visibility": "true",
				"resource-tags":          `{invalid}`,
			},
			expectedError: "failed to parse resource tags JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("resource-tags", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagPortInput(cmd)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessFlagLAGPortInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "valid LAG input",
			flags: map[string]string{
				"name":                   "test-lag",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "1",
				"lag-count":              "2",
				"marketplace-visibility": "true",
			},
		},
		{
			name: "invalid resource tags JSON",
			flags: map[string]string{
				"name":                   "test-lag",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "1",
				"lag-count":              "2",
				"marketplace-visibility": "true",
				"resource-tags":          `{bad json}`,
			},
			expectedError: "failed to parse resource tags JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Int("lag-count", 0, "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("resource-tags", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagLAGPortInput(cmd)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessJSONPortInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"name":"test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true}`,
		},
		{
			name:          "invalid JSON string",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path/config.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true}`,
		},
		{
			name:          "empty tag key rejected",
			jsonStr:       `{"name":"test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true,"resourceTags":{"":"x"}}`,
			expectedError: "tag key must not be empty",
		},
		{
			name:          "empty tag key rejected via file",
			writeFile:     `{"name":"test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true,"resourceTags":{"":"x"}}`,
			expectedError: "tag key must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "port-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONPortInput(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessJSONPortInput_ValidResourceTags(t *testing.T) {
	req, err := processJSONPortInput(`{"name":"test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true,"resourceTags":{"env":"prod"}}`, "")
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "prod", req.ResourceTags["env"])
}

func TestProcessJSONLAGPortInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
		expectedLag   int
	}{
		{
			name:        "valid LAG JSON",
			jsonStr:     `{"name":"lag-test","term":12,"portSpeed":10000,"locationId":1,"lagCount":4,"marketPlaceVisibility":true}`,
			expectedLag: 4,
		},
		{
			name:          "missing lagCount rejected",
			jsonStr:       `{"name":"lag-test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true}`,
			expectedError: "LAG count",
		},
		{
			name:          "zero lagCount rejected",
			jsonStr:       `{"name":"lag-test","term":12,"portSpeed":10000,"locationId":1,"lagCount":0,"marketPlaceVisibility":true}`,
			expectedError: "LAG count",
		},
		{
			name:          "lagCount above max rejected",
			jsonStr:       `{"name":"lag-test","term":12,"portSpeed":10000,"locationId":1,"lagCount":9,"marketPlaceVisibility":true}`,
			expectedError: "LAG count",
		},
		{
			name:          "non-LAG port speed rejected",
			jsonStr:       `{"name":"lag-test","term":12,"portSpeed":1000,"locationId":1,"lagCount":4,"marketPlaceVisibility":true}`,
			expectedError: "port speed",
		},
		{
			name:          "invalid JSON string",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "empty tag key rejected",
			jsonStr:       `{"name":"lag-test","term":12,"portSpeed":10000,"locationId":1,"lagCount":4,"marketPlaceVisibility":true,"resourceTags":{"":"x"}}`,
			expectedError: "tag key must not be empty",
		},
		{
			name:        "valid LAG JSON file",
			writeFile:   `{"name":"lag-file-test","term":12,"portSpeed":10000,"locationId":1,"lagCount":4,"marketPlaceVisibility":true}`,
			expectedLag: 4,
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path/lag.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:          "missing lagCount rejected via file",
			writeFile:     `{"name":"lag-file-test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true}`,
			expectedError: "LAG count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "lag-port-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONLAGPortInput(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)
				assert.Equal(t, tt.expectedLag, req.LagCount)
			}
		})
	}
}

func TestProcessJSONUpdatePortInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid update JSON",
			jsonStr: `{"name":"updated-port"}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "no fields updated",
			jsonStr:       `{}`,
			expectedError: "at least one field must be updated",
		},
		{
			name:          "Invalid contract term",
			jsonStr:       `{"contractTermMonths":99}`,
			expectedError: "Invalid contract term",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path/config.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-update"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "port-update-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONUpdatePortInput(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessFlagUpdatePortInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
		checkReq      func(*testing.T, *cobra.Command)
	}{
		{
			name:          "no flags changed",
			flags:         map[string]string{},
			expectedError: "at least one field must be updated",
		},
		{
			name:  "name only",
			flags: map[string]string{"name": "new-name"},
		},
		{
			name:  "marketplace-visibility only",
			flags: map[string]string{"marketplace-visibility": "true"},
		},
		{
			name:  "cost-centre only",
			flags: map[string]string{"cost-centre": "IT-2024"},
		},
		{
			name:  "term only",
			flags: map[string]string{"term": "24"},
		},
		{
			name:          "invalid term",
			flags:         map[string]string{"term": "99"},
			expectedError: "Invalid contract term",
		},
		{
			name: "multiple fields",
			flags: map[string]string{
				"name":        "new-name",
				"cost-centre": "IT-2024",
				"term":        "12",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Int("term", 0, "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagUpdatePortInput(cmd, "port-uid-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "port-uid-123", req.PortID)
			}
		})
	}
}
