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

// Login logs into the Megaport API using the current profile or environment variables.
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	// Priority for credentials:
	// 1. Active profile
	// 2. Environment variables (fallback)

	var accessKey, secretKey, env string

	// First try to use the active profile
	manager, err := NewConfigManager()
	if err == nil { // Only try profile if config can be loaded
		profile, _, err := manager.GetCurrentProfile()
		if err == nil { // Only use profile if active profile exists
			// Use credentials from profile
			accessKey = profile.AccessKey
			secretKey = profile.SecretKey
			env = profile.Environment
		}
	}

	// If no active profile or profile credentials are incomplete, fall back to environment variables
	if accessKey == "" {
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
	}
	if env == "" {
		env = os.Getenv("MEGAPORT_ENVIRONMENT")
	}

	// Validate credentials
	if accessKey == "" {
		return nil, fmt.Errorf("megaport API access key not provided. Configure an active profile or set MEGAPORT_ACCESS_KEY environment variable")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("megaport API secret key not provided. Configure an active profile or set MEGAPORT_SECRET_KEY environment variable")
	}

	// Default to production environment if not specified
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

	// Create client without authenticating yet
	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(accessKey, secretKey), envOpt)
	if err != nil {
		return nil, err
	}

	// Start spinner for login
	spinner := output.PrintLoggingIn(false) // You might want to pass a noColor flag here if available

	// Authenticate with the API
	_, err = megaportClient.Authorize(ctx)

	// Stop spinner, handling success or error
	if err != nil {
		spinner.Stop() // Just stop the spinner without success message if there's an error
		return nil, err
	} else {
		spinner.StopWithSuccess("Successfully logged in to Megaport")
	}

	return megaportClient, nil
}
