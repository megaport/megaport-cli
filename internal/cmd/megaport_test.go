package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	// Save original login function to restore after tests
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		envFlag     string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "all valid env vars",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: false,
		},
		{
			name: "empty access key",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name: "empty secret key",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name:        "no env vars set",
			envVars:     map[string]string{},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name: "flag overrides env var",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			envFlag:     "production",
			shouldError: false,
		},
		{
			name: "flag provides env when not set",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY": "test-access-key",
				"MEGAPORT_SECRET_KEY": "test-secret-key",
			},
			envFlag:     "production",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables before each test
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
			os.Unsetenv("MEGAPORT_SECRET_KEY")
			os.Unsetenv("MEGAPORT_ENVIRONMENT")

			// Set environment variables for this test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Set the environment flag if provided
			if tt.envFlag != "" {
				env = tt.envFlag
			} else {
				env = ""
			}

			// Mock the loginFunc to capture inputs and return results based on test case
			var capturedAccessKey, capturedSecretKey, capturedEnv string

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				capturedAccessKey = os.Getenv(accessKeyEnvVar)
				capturedSecretKey = os.Getenv(secretKeyEnvVar)
				capturedEnv = env
				if capturedEnv == "" {
					capturedEnv = os.Getenv(environmentEnvVar)
				}

				if capturedAccessKey == "" || capturedSecretKey == "" {
					return nil, fmt.Errorf("access key, secret key, and environment are required")
				}

				client := &megaport.Client{}
				return client, nil
			}

			// Call the Login function
			client, err := Login(context.Background())

			// Verify results
			if tt.shouldError {
				assert.Nil(t, client)
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NotNil(t, client)
				assert.NoError(t, err)

				// If test succeeds, verify the environment was properly set
				if tt.envFlag != "" {
					assert.Equal(t, tt.envFlag, capturedEnv)
				} else if envVal, ok := tt.envVars["MEGAPORT_ENVIRONMENT"]; ok {
					assert.Equal(t, envVal, capturedEnv)
				} else {
					// Default should be production
					assert.Equal(t, "production", capturedEnv)
				}
			}
		})
	}
}
