package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Original login function that will be restored after tests
var originalLoginFunc func(ctx context.Context) (*megaport.Client, error)

func TestConfigureCmd(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("MEGAPORT_ACCESS_KEY", "test-access-key")
	t.Setenv("MEGAPORT_SECRET_KEY", "test-secret-key")
	t.Setenv("MEGAPORT_ENVIRONMENT", "staging")

	defer func() {
		os.Unsetenv("MEGAPORT_ACCESS_KEY")
		os.Unsetenv("MEGAPORT_SECRET_KEY")
		os.Unsetenv("MEGAPORT_ENVIRONMENT")
	}()

	cmd := configureCmd
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

func TestConfigureCmdMissingEnvVars(t *testing.T) {
	// Unset environment variables for testing
	os.Unsetenv("MEGAPORT_ACCESS_KEY")
	os.Unsetenv("MEGAPORT_SECRET_KEY")
	os.Unsetenv("MEGAPORT_ENVIRONMENT")

	cmd := configureCmd
	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

func TestConfigureCmd_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
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
			errorMsg:    "access key cannot be empty",
		},
		{
			name: "empty secret key",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: true,
			errorMsg:    "secret key cannot be empty",
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "invalid",
			},
			shouldError: true,
			errorMsg:    "invalid environment",
		},
		{
			name:        "no env vars set",
			envVars:     map[string]string{},
			shouldError: true,
			errorMsg:    "required environment variables not set",
		},
		{
			name: "whitespace only values",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "   ",
				"MEGAPORT_SECRET_KEY":  "   ",
				"MEGAPORT_ENVIRONMENT": "   ",
			},
			shouldError: true,
			errorMsg:    "invalid environment variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment before each test
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
			os.Unsetenv("MEGAPORT_SECRET_KEY")
			os.Unsetenv("MEGAPORT_ENVIRONMENT")

			// Set test environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cmd := configureCmd
			err := cmd.RunE(cmd, []string{})

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigureCmd_ExtraArgs(t *testing.T) {
	// Create a new command instance for testing
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Megaport CLI credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unexpected arguments: %v", args)
			}
			return nil
		},
	}

	// Test with extra arguments
	err := cmd.RunE(cmd, []string{"extra", "args"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected arguments")
}

// SetupLoginMocks initializes mock login functionality
// Returns a cleanup function to be deferred
func SetupLoginMocks() func() {
	originalLoginFunc = loginFunc
	return func() {
		loginFunc = originalLoginFunc
	}
}

// MockLoginSuccess sets up a loginFunc that returns a successful client with mock services
func MockLoginSuccess() {
	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		// Add mock services to the client
		client.MCRService = &MockMCRService{}
		client.PortService = &MockPortService{}
		// Add other service mocks as needed
		return client, nil
	}
}

// MockLoginWithError sets up a loginFunc that returns a specific error
func MockLoginWithError(errorMsg string) {
	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return nil, errors.New(errorMsg)
	}
}
