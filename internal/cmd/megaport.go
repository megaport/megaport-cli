package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// Configuration file path
var (
	accessKeyEnvVar   = "MEGAPORT_ACCESS_KEY"
	secretKeyEnvVar   = "MEGAPORT_SECRET_KEY"
	environmentEnvVar = "MEGAPORT_ENVIRONMENT"
)

// Login mocks client auth (actual API calls are not performed in tests)
func Login(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}
	accessKey := os.Getenv(accessKeyEnvVar)
	secretKey := os.Getenv(secretKeyEnvVar)
	environment := os.Getenv(environmentEnvVar)

	if accessKey == "" || secretKey == "" || environment == "" {
		fmt.Println("Please provide access key and secret key using environment variables MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, and MEGAPORT_ENVIRONMENT")
		return nil, fmt.Errorf("access key, secret key, and environment are required")
	}

	var envOpt megaport.ClientOpt
	switch environment {
	case "production":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(accessKey, secretKey), envOpt)
	if err != nil {
		return nil, err
	}
	if _, err := megaportClient.Authorize(ctx); err != nil {
		return nil, err
	}
	return megaportClient, nil
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI with your credentials",
	Long: `Configure the CLI with your Megaport API credentials.

You must provide credentials through environment variables:
  MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, and MEGAPORT_ENVIRONMENT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}

		accessKey := os.Getenv(accessKeyEnvVar)
		secretKey := os.Getenv(secretKeyEnvVar)
		environment := os.Getenv(environmentEnvVar)

		// Check if any env vars are not set
		if accessKey == "" && secretKey == "" && environment == "" {
			return fmt.Errorf("required environment variables not set")
		}

		// Check individual env vars
		if accessKey == "" {
			return fmt.Errorf("access key cannot be empty")
		}
		if secretKey == "" {
			return fmt.Errorf("secret key cannot be empty")
		}
		if environment == "" {
			return fmt.Errorf("environment cannot be empty")
		}

		// Check for whitespace-only values
		if strings.TrimSpace(accessKey) == "" ||
			strings.TrimSpace(secretKey) == "" ||
			strings.TrimSpace(environment) == "" {
			return fmt.Errorf("invalid environment variables")
		}

		// Validate environment
		switch strings.TrimSpace(environment) {
		case "production", "staging", "development":
			// valid
		default:
			return fmt.Errorf("invalid environment: %s", environment)
		}

		fmt.Printf("Environment (%s) configured successfully.\n", environment)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
