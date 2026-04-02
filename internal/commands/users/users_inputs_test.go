package users

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessJSONCreateUserInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"firstName":"John","lastName":"Doe","email":"john@example.com","position":"Technical Admin","active":true}`,
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
			writeFile: `{"firstName":"Jane","lastName":"Doe","email":"jane@example.com","position":"Finance","active":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "user-create-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONCreateUserInput(tt.jsonStr, jsonFile)
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

func TestProcessJSONUpdateUserInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid update",
			jsonStr: `{"firstName":"Updated"}`,
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
			writeFile: `{"lastName":"NewLast"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "user-update-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONUpdateUserInput(tt.jsonStr, jsonFile)
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

func TestProcessFlagCreateUserInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "valid input",
			flags: map[string]string{
				"first-name": "John",
				"last-name":  "Doe",
				"email":      "john@example.com",
				"position":   "Technical Admin",
			},
		},
		{
			name: "with phone",
			flags: map[string]string{
				"first-name": "John",
				"last-name":  "Doe",
				"email":      "john@example.com",
				"position":   "Finance",
				"phone":      "+61412345678",
			},
		},
		{
			name: "invalid position",
			flags: map[string]string{
				"first-name": "John",
				"last-name":  "Doe",
				"email":      "john@example.com",
				"position":   "Super Admin",
			},
			expectedError: "invalid position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("first-name", "", "")
			cmd.Flags().String("last-name", "", "")
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("position", "", "")
			cmd.Flags().String("phone", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagCreateUserInput(cmd)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.True(t, req.Active)
			}
		})
	}
}

func TestProcessFlagUpdateUserInput(t *testing.T) {
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
			name:  "first-name only",
			flags: map[string]string{"first-name": "NewFirst"},
		},
		{
			name:  "email only",
			flags: map[string]string{"email": "new@example.com"},
		},
		{
			name:  "active flag",
			flags: map[string]string{"active": "true"},
		},
		{
			name:  "notification-enabled flag",
			flags: map[string]string{"notification-enabled": "true"},
		},
		{
			name: "multiple fields",
			flags: map[string]string{
				"first-name": "New",
				"last-name":  "Name",
				"position":   "Finance",
				"phone":      "+61400000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("first-name", "", "")
			cmd.Flags().String("last-name", "", "")
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("position", "", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().Bool("active", false, "")
			cmd.Flags().Bool("notification-enabled", false, "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagUpdateUserInput(cmd)
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
