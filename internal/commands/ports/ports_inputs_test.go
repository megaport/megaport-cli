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
			expectedError: "error parsing resource tags JSON",
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
			expectedError: "error parsing resource tags JSON",
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
			expectedError: "error parsing JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path/config.json",
			expectedError: "error reading JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-test","term":12,"portSpeed":10000,"locationId":1,"marketPlaceVisibility":true}`,
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
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
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
			expectedError: "error parsing JSON",
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
			expectedError: "error reading JSON file",
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
