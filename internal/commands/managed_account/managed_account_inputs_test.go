package managed_account

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildManagedAccountRequestFromJSON_Variants(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedName  string
		expectedRef   string
		expectedError string
	}{
		{
			name:         "valid JSON string",
			jsonStr:      `{"accountName":"Test","accountRef":"REF"}`,
			expectedName: "Test",
			expectedRef:  "REF",
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "empty JSON string with no file",
			jsonStr:       "",
			expectedError: "failed to parse JSON",
		},
		{
			name:         "valid JSON file",
			writeFile:    `{"accountName":"FileAccount","accountRef":"FILE-REF"}`,
			expectedName: "FileAccount",
			expectedRef:  "FILE-REF",
		},
		{
			name:          "missing file",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "managed-account-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := buildManagedAccountRequestFromJSON(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, tt.expectedName, req.AccountName)
				assert.Equal(t, tt.expectedRef, req.AccountRef)
			}
		})
	}
}

func TestBuildManagedAccountRequestFromFlags_Variants(t *testing.T) {
	tests := []struct {
		name         string
		flags        map[string]string
		expectedName string
		expectedRef  string
	}{
		{
			name: "both flags set",
			flags: map[string]string{
				"account-name": "TestAccount",
				"account-ref":  "REF-001",
			},
			expectedName: "TestAccount",
			expectedRef:  "REF-001",
		},
		{
			name: "name only",
			flags: map[string]string{
				"account-name": "TestAccount",
			},
			expectedName: "TestAccount",
			expectedRef:  "",
		},
		{
			name:         "no flags set",
			flags:        map[string]string{},
			expectedName: "",
			expectedRef:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := buildManagedAccountRequestFromFlags(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			assert.Equal(t, tt.expectedName, req.AccountName)
			assert.Equal(t, tt.expectedRef, req.AccountRef)
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromJSON_Variants(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedName  string
		expectedRef   string
		expectedError string
	}{
		{
			name:         "valid update JSON",
			jsonStr:      `{"accountName":"Updated","accountRef":"NEW-REF"}`,
			expectedName: "Updated",
			expectedRef:  "NEW-REF",
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:         "valid JSON file",
			writeFile:    `{"accountName":"FileUpdate","accountRef":"FILE-REF"}`,
			expectedName: "FileUpdate",
			expectedRef:  "FILE-REF",
		},
		{
			name:          "missing file",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "managed-account-update-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := buildUpdateManagedAccountRequestFromJSON(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, tt.expectedName, req.AccountName)
				assert.Equal(t, tt.expectedRef, req.AccountRef)
			}
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromFlags_Variants(t *testing.T) {
	tests := []struct {
		name         string
		flags        map[string]string
		expectedName string
		expectedRef  string
	}{
		{
			name: "name changed only",
			flags: map[string]string{
				"account-name": "NewName",
			},
			expectedName: "NewName",
			expectedRef:  "",
		},
		{
			name: "both changed",
			flags: map[string]string{
				"account-name": "NewName",
				"account-ref":  "NewRef",
			},
			expectedName: "NewName",
			expectedRef:  "NewRef",
		},
		{
			name:         "no flags changed",
			flags:        map[string]string{},
			expectedName: "",
			expectedRef:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := buildUpdateManagedAccountRequestFromFlags(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			assert.Equal(t, tt.expectedName, req.AccountName)
			assert.Equal(t, tt.expectedRef, req.AccountRef)
		})
	}
}
