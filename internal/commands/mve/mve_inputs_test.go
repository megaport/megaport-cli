package mve

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVendorConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
	}{
		{
			name:          "missing vendor field",
			config:        map[string]interface{}{"imageId": float64(1)},
			expectedError: "vendor field is required",
		},
		{
			name:          "unsupported vendor",
			config:        map[string]interface{}{"vendor": "unknown"},
			expectedError: "unsupported vendor",
		},
		{
			name: "6wind success",
			config: map[string]interface{}{
				"vendor": "6wind", "imageId": float64(1), "productSize": "MEDIUM",
				"sshPublicKey": "ssh-rsa AAAA",
			},
		},
		{
			name: "aruba success",
			config: map[string]interface{}{
				"vendor": "aruba", "imageId": float64(1), "productSize": "MEDIUM",
				"accountName": "acct", "accountKey": "key",
			},
		},
		{
			name: "aviatrix success",
			config: map[string]interface{}{
				"vendor": "aviatrix", "imageId": float64(1), "productSize": "MEDIUM",
				"cloudInit": "#cloud-config",
			},
		},
		{
			name: "cisco success",
			config: map[string]interface{}{
				"vendor": "cisco", "imageId": float64(1), "productSize": "MEDIUM",
				"mveLabel": "label", "manageLocally": true,
				"adminSshPublicKey": "ssh-rsa", "sshPublicKey": "ssh-rsa",
				"cloudInit": "#cloud", "fmcIpAddress": "10.0.0.1",
				"fmcRegistrationKey": "key", "fmcNatId": "nat",
			},
		},
		{
			name: "fortinet success",
			config: map[string]interface{}{
				"vendor": "fortinet", "imageId": float64(1), "productSize": "MEDIUM",
				"adminSshPublicKey": "ssh-rsa", "sshPublicKey": "ssh-rsa",
				"licenseData": "license",
			},
		},
		{
			name: "palo_alto success",
			config: map[string]interface{}{
				"vendor": "palo_alto", "imageId": float64(1), "productSize": "MEDIUM",
				"sshPublicKey": "ssh-rsa", "adminPasswordHash": "hash",
				"licenseData": "license",
			},
		},
		{
			name: "prisma success",
			config: map[string]interface{}{
				"vendor": "prisma", "imageId": float64(1), "productSize": "MEDIUM",
				"ionKey": "ion", "secretKey": "secret",
			},
		},
		{
			name: "versa success",
			config: map[string]interface{}{
				"vendor": "versa", "imageId": float64(1), "productSize": "MEDIUM",
				"directorAddress": "dir", "controllerAddress": "ctrl",
				"localAuth": "local", "remoteAuth": "remote", "serialNumber": "SN123",
			},
		},
		{
			name: "vmware success",
			config: map[string]interface{}{
				"vendor": "vmware", "imageId": float64(1), "productSize": "MEDIUM",
				"adminSshPublicKey": "ssh-rsa", "sshPublicKey": "ssh-rsa",
				"vcoAddress": "vco.example.com", "vcoActivationCode": "code",
			},
		},
		{
			name: "meraki success",
			config: map[string]interface{}{
				"vendor": "meraki", "imageId": float64(1), "productSize": "MEDIUM",
				"token": "token123",
			},
		},
		// Missing required field tests
		{
			name:          "6wind missing sshPublicKey",
			config:        map[string]interface{}{"vendor": "6wind", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "sshPublicKey is required",
		},
		{
			name:          "aruba missing accountName",
			config:        map[string]interface{}{"vendor": "aruba", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "accountName is required",
		},
		{
			name:          "aviatrix missing cloudInit",
			config:        map[string]interface{}{"vendor": "aviatrix", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "cloudInit is required",
		},
		{
			name:          "fortinet missing adminSshPublicKey",
			config:        map[string]interface{}{"vendor": "fortinet", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "adminSshPublicKey is required",
		},
		{
			name:          "palo_alto missing sshPublicKey",
			config:        map[string]interface{}{"vendor": "palo_alto", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "sshPublicKey is required",
		},
		{
			name:          "prisma missing ionKey",
			config:        map[string]interface{}{"vendor": "prisma", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "ionKey is required",
		},
		{
			name:          "versa missing directorAddress",
			config:        map[string]interface{}{"vendor": "versa", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "directorAddress is required",
		},
		{
			name:          "vmware missing adminSshPublicKey",
			config:        map[string]interface{}{"vendor": "vmware", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "adminSshPublicKey is required",
		},
		{
			name:          "meraki missing token",
			config:        map[string]interface{}{"vendor": "meraki", "imageId": float64(1), "productSize": "MEDIUM"},
			expectedError: "token is required",
		},
		{
			name:          "missing imageId",
			config:        map[string]interface{}{"vendor": "cisco", "productSize": "MEDIUM"},
			expectedError: "imageId is required",
		},
		{
			name:          "missing productSize",
			config:        map[string]interface{}{"vendor": "cisco", "imageId": float64(1)},
			expectedError: "productSize is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVendorConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestProcessJSONUpdateMVEInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid name update",
			jsonStr: `{"name":"updated-mve"}`,
		},
		{
			name:    "valid cost centre update",
			jsonStr: `{"costCentre":"IT-2024"}`,
		},
		{
			name:    "valid term update",
			jsonStr: `{"contractTermMonths":24}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "error parsing JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
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
				tmpFile, err := os.CreateTemp("", "mve-update-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONUpdateMVEInput(tt.jsonStr, jsonFile, "mve-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "mve-123", req.MVEID)
			}
		})
	}
}

func TestProcessFlagUpdateMVEInput(t *testing.T) {
	tests := []struct {
		name          string
		nameVal       string
		costCentre    string
		contractTerm  int
		expectedError string
	}{
		{
			name:    "name update",
			nameVal: "new-mve",
		},
		{
			name:       "cost centre update",
			costCentre: "IT-2024",
		},
		{
			name:         "contract term update",
			contractTerm: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCmd()
			if tt.nameVal != "" {
				require.NoError(t, cmd.Flags().Set("name", tt.nameVal))
			}
			if tt.costCentre != "" {
				require.NoError(t, cmd.Flags().Set("cost-centre", tt.costCentre))
			}
			if tt.contractTerm > 0 {
				require.NoError(t, cmd.Flags().Set("contract-term", "24"))
			}

			req, err := processFlagUpdateMVEInput(cmd, "mve-123")
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

func createTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Int("contract-term", 0, "")
	return cmd
}
