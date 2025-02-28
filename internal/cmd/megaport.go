package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	megaport "github.com/megaport/megaportgo"
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
	environment := env
	if environment == "" {
		environment = os.Getenv(environmentEnvVar)
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "production", "Environment to use (production, staging, development)")
}
