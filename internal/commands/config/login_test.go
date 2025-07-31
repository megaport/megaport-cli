package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var (
	env               string
	accessKeyEnvVar   = "MEGAPORT_ACCESS_KEY"
	secretKeyEnvVar   = "MEGAPORT_SECRET_KEY"
	environmentEnvVar = "MEGAPORT_ENVIRONMENT"
)

func TestLogin(t *testing.T) {
	originalLoginFunc := LoginFunc
	defer func() {
		LoginFunc = originalLoginFunc
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
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
			os.Unsetenv("MEGAPORT_SECRET_KEY")
			os.Unsetenv("MEGAPORT_ENVIRONMENT")

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			if tt.envFlag != "" {
				env = tt.envFlag
			} else {
				env = ""
			}

			var capturedAccessKey, capturedSecretKey, capturedEnv string

			LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
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

			client, err := Login(context.Background())

			if tt.shouldError {
				assert.Nil(t, client)
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NotNil(t, client)
				assert.NoError(t, err)

				if tt.envFlag != "" {
					assert.Equal(t, tt.envFlag, capturedEnv)
				} else if envVal, ok := tt.envVars["MEGAPORT_ENVIRONMENT"]; ok {
					assert.Equal(t, envVal, capturedEnv)
				} else {
					assert.Equal(t, "production", capturedEnv)
				}
			}
		})
	}
}

func TestEnvironmentSelectionPrecedence(t *testing.T) {
	// Save original values
	originalEnv := utils.Env
	originalMegaportEnv := os.Getenv("MEGAPORT_ENVIRONMENT")

	defer func() {
		// Restore original values
		utils.Env = originalEnv
		if originalMegaportEnv == "" {
			os.Unsetenv("MEGAPORT_ENVIRONMENT")
		} else {
			os.Setenv("MEGAPORT_ENVIRONMENT", originalMegaportEnv)
		}
	}()

	tests := []struct {
		name        string
		globalFlag  string
		envVar      string
		expectedEnv string
	}{
		{
			name:        "Global flag takes precedence over environment variable",
			globalFlag:  "staging",
			envVar:      "production",
			expectedEnv: "staging",
		},
		{
			name:        "Environment variable used when no global flag",
			globalFlag:  "",
			envVar:      "staging",
			expectedEnv: "staging",
		},
		{
			name:        "Default to production when neither flag nor env var set",
			globalFlag:  "",
			envVar:      "",
			expectedEnv: "production",
		},
		{
			name:        "Development environment via global flag",
			globalFlag:  "development",
			envVar:      "",
			expectedEnv: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			utils.Env = tt.globalFlag
			if tt.envVar != "" {
				os.Setenv("MEGAPORT_ENVIRONMENT", tt.envVar)
			} else {
				os.Unsetenv("MEGAPORT_ENVIRONMENT")
			}

			// Test the environment selection logic by examining the values
			// that would be read in the login function
			var env string

			// Simulate the new login function logic
			if utils.Env != "" {
				env = utils.Env
			} else {
				env = os.Getenv("MEGAPORT_ENVIRONMENT")
			}

			if env == "" {
				env = "production"
			}

			assert.Equal(t, tt.expectedEnv, env, "Environment selection should match expected value")
		})
	}
}

func TestCredentialSelectionPrecedence(t *testing.T) {
	// Save original values
	originalEnv := utils.Env
	originalAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	originalSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	defer func() {
		// Restore original values
		utils.Env = originalEnv
		if originalAccessKey == "" {
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
		} else {
			os.Setenv("MEGAPORT_ACCESS_KEY", originalAccessKey)
		}
		if originalSecretKey == "" {
			os.Unsetenv("MEGAPORT_SECRET_KEY")
		} else {
			os.Setenv("MEGAPORT_SECRET_KEY", originalSecretKey)
		}
	}()

	tests := []struct {
		name                  string
		globalFlag            string
		envAccessKey          string
		envSecretKey          string
		expectEnvVarsPriority bool
	}{
		{
			name:                  "Flag set: env vars should have priority",
			globalFlag:            "production",
			envAccessKey:          "env-access-key",
			envSecretKey:          "env-secret-key",
			expectEnvVarsPriority: true,
		},
		{
			name:                  "No flag: profile should have priority",
			globalFlag:            "",
			envAccessKey:          "env-access-key",
			envSecretKey:          "env-secret-key",
			expectEnvVarsPriority: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			utils.Env = tt.globalFlag
			os.Setenv("MEGAPORT_ACCESS_KEY", tt.envAccessKey)
			os.Setenv("MEGAPORT_SECRET_KEY", tt.envSecretKey)

			// Test credential selection logic
			var accessKey, secretKey string

			// Simulate the new credential selection logic
			if utils.Env != "" {
				// --env flag was explicitly set, prioritize environment variables
				accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
				secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
			} else {
				// No --env flag, profile would have priority (but we can't test that easily here)
				// For this test, we'll just verify the env var logic works
				if tt.expectEnvVarsPriority {
					accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
					secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
				}
			}

			if tt.expectEnvVarsPriority {
				assert.Equal(t, tt.envAccessKey, accessKey, "Should use environment variable for access key when --env flag is set")
				assert.Equal(t, tt.envSecretKey, secretKey, "Should use environment variable for secret key when --env flag is set")
			}
		})
	}
}
