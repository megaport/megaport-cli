package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	if environment == "" {
		environment = env
	}

	if accessKey == "" || secretKey == "" {
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
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
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

var (
	version = "dev" // This will be overridden by the build process
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Megaport CLI",
	Long:  `All software has versions. This is Megaport CLI's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Megaport CLI Version:", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "production", "Environment to use (production, staging, development)")
}
