package config

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

func Login(ctx context.Context) (*megaport.Client, error) {
	return LoginFunc(ctx)
}

// LoginFunc logs into the Megaport API using the current profile or environment variables.
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	var accessKey, secretKey, env string

	manager, err := NewConfigManager()
	if err == nil {
		profile, _, err := manager.GetCurrentProfile()
		if err == nil {
			accessKey = profile.AccessKey
			secretKey = profile.SecretKey
			env = profile.Environment
		}
	}

	if accessKey == "" {
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
	}
	if env == "" {
		env = os.Getenv("MEGAPORT_ENVIRONMENT")
	}

	if accessKey == "" {
		return nil, fmt.Errorf("megaport API access key not provided. Configure an active profile or set MEGAPORT_ACCESS_KEY environment variable")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("megaport API secret key not provided. Configure an active profile or set MEGAPORT_SECRET_KEY environment variable")
	}

	if env == "" {
		env = "production"
	}

	var envOpt megaport.ClientOpt
	switch env {
	case "production":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	}

	httpClient := &http.Client{}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(accessKey, secretKey), envOpt)
	if err != nil {
		return nil, err
	}

	spinner := output.PrintLoggingIn(false)
	_, err = megaportClient.Authorize(ctx)

	if err != nil {
		spinner.Stop()
		return nil, err
	} else {
		spinner.StopWithSuccess("Successfully logged in to Megaport")
	}

	return megaportClient, nil
}
