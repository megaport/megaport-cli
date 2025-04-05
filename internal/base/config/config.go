package config

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

var (
	accessKeyEnvVar   = "MEGAPORT_ACCESS_KEY"
	secretKeyEnvVar   = "MEGAPORT_SECRET_KEY"
	environmentEnvVar = "MEGAPORT_ENVIRONMENT"
)

var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}
	accessKey := os.Getenv(accessKeyEnvVar)
	secretKey := os.Getenv(secretKeyEnvVar)
	environment := os.Getenv(environmentEnvVar)
	if environment == "" {
		environment = utils.GetCurrentEnv()
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
	return LoginFunc(ctx)
}
