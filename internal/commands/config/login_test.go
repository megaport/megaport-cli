package config

import (
	"context"
	"fmt"
	"os"
	"testing"

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
