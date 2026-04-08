package mcr

import (
	"context"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessJSONMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-mcr","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "mcr-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONMCRInput(tt.jsonStr, jsonFile)
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

func TestProcessJSONUpdateMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid update",
			jsonStr: `{"name":"updated-mcr"}`,
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
			name:          "invalid contract term",
			jsonStr:       `{"contractTermMonths":99}`,
			expectedError: "Invalid contract term",
		},
		{
			name:          "empty name provided",
			jsonStr:       `{"name":""}`,
			expectedError: "name cannot be empty",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
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
				tmpFile, err := os.CreateTemp("", "mcr-update-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONUpdateMCRInput(tt.jsonStr, jsonFile)
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

func TestProcessFlagUpdateMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
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
			name:  "cost-centre only",
			flags: map[string]string{"cost-centre": "IT-2024"},
		},
		{
			name:  "marketplace-visibility only",
			flags: map[string]string{"marketplace-visibility": "true"},
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
			name:          "empty name",
			flags:         map[string]string{"name": ""},
			expectedError: "name cannot be empty",
		},
		{
			name:  "multiple fields",
			flags: map[string]string{"name": "new", "cost-centre": "IT", "term": "12"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().Int("term", 0, "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagUpdateMCRInput(cmd, "mcr-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "mcr-123", req.MCRID)
			}
		})
	}
}

func TestProcessJSONPrefixFilterListInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid",
			jsonStr: `{"description":"PFL","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8"}]}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"description":"File PFL","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "pfl-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONPrefixFilterListInput(tt.jsonStr, jsonFile, "mcr-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "mcr-123", req.MCRID)
			}
		})
	}
}

func TestProcessFlagPrefixFilterListInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "valid",
			flags: map[string]string{
				"description":    "My PFL",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
		},
		{
			name: "invalid entries JSON",
			flags: map[string]string{
				"description":    "My PFL",
				"address-family": "IPv4",
				"entries":        `{invalid}`,
			},
			expectedError: "failed to parse entries JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagPrefixFilterListInput(cmd, "mcr-123")
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

func TestProcessFlagUpdatePrefixFilterListInput(t *testing.T) {
	// Setup mock for getMCRPrefixFilterListFunc
	originalGetPFL := getMCRPrefixFilterListFunc
	originalLogin := config.GetLoginFunc()
	defer func() {
		getMCRPrefixFilterListFunc = originalGetPFL
		config.SetLoginFunc(originalLogin)
	}()

	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	})
	getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, pflID int) (*megaport.MCRPrefixFilterList, error) {
		return &megaport.MCRPrefixFilterList{
			ID:            1,
			Description:   "Current PFL",
			AddressFamily: "IPv4",
			Entries: []*megaport.MCRPrefixListEntry{
				{Action: "permit", Prefix: "10.0.0.0/8"},
			},
		}, nil
	}

	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name:          "no flags changed",
			flags:         map[string]string{},
			expectedError: "at least one field",
		},
		{
			name:  "description only",
			flags: map[string]string{"description": "Updated PFL"},
		},
		{
			name:  "entries only",
			flags: map[string]string{"entries": `[{"action":"deny","prefix":"192.168.0.0/16"}]`},
		},
		{
			name:  "both description and entries",
			flags: map[string]string{"description": "New PFL", "entries": `[{"action":"permit","prefix":"172.16.0.0/12"}]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			pfl, err := processFlagUpdatePrefixFilterListInput(cmd, "mcr-123", 1)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pfl)
				assert.Equal(t, 1, pfl.ID)
			}
		})
	}
}

func TestProcessFlagMCRInput_ResourceTags(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "with valid resource tags",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
				"resource-tags":          `{"env":"prod"}`,
			},
		},
		{
			name: "with invalid resource tags JSON",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
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
			cmd.Flags().Int("mcr-asn", 0, "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("resource-tags", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagMCRInput(cmd)
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
