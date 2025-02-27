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

var (
	accessKeyEnvVar   = "MEGAPORT_ACCESS_KEY"
	secretKeyEnvVar   = "MEGAPORT_SECRET_KEY"
	environmentEnvVar = "MEGAPORT_ENVIRONMENT"
)

var loginFunc = func(ctx context.Context) (*megaport.Client, error) {
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

func Login(ctx context.Context) (*megaport.Client, error) {
	return loginFunc(ctx)
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

		if accessKey == "" && secretKey == "" && environment == "" {
			return fmt.Errorf("required environment variables not set")
		}

		if accessKey == "" {
			return fmt.Errorf("access key cannot be empty")
		}
		if secretKey == "" {
			return fmt.Errorf("secret key cannot be empty")
		}
		if environment == "" {
			return fmt.Errorf("environment cannot be empty")
		}

		if strings.TrimSpace(accessKey) == "" ||
			strings.TrimSpace(secretKey) == "" ||
			strings.TrimSpace(environment) == "" {
			return fmt.Errorf("invalid environment variables")
		}

		switch strings.TrimSpace(environment) {
		case "production", "staging", "development":
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
