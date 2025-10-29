// go:build !js && !wasm
//go:build !js && !wasm
// +build !js,!wasm

package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func Login(ctx context.Context) (*megaport.Client, error) {
	return LoginFunc(ctx)
}

func LoginWithOutput(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	return LoginFuncWithOutput(ctx, outputFormat)
}

// LoginFunc logs into the Megaport API using the current profile or environment variables.
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	var accessKey, secretKey, env string

	// Environment selection priority: --env flag > profile > env var > default
	if utils.Env != "" {
		env = utils.Env
	} else {
		// Check profile environment
		manager, err := NewConfigManager()
		if err == nil {
			profile, _, err := manager.GetCurrentProfile()
			if err == nil && profile.Environment != "" {
				env = profile.Environment
			}
		}

		// Fall back to environment variable if still not set
		if env == "" {
			env = os.Getenv("MEGAPORT_ENVIRONMENT")
		}
	}

	if env == "" {
		env = "production"
	}

	// Credential selection: if --env flag is used, prefer env vars over profile
	if utils.Env != "" {
		// --env flag was explicitly set, prioritize environment variables
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

		// If env vars are empty, fall back to profile
		if accessKey == "" || secretKey == "" {
			manager, err := NewConfigManager()
			if err == nil {
				profile, _, err := manager.GetCurrentProfile()
				if err == nil {
					if accessKey == "" {
						accessKey = profile.AccessKey
					}
					if secretKey == "" {
						secretKey = profile.SecretKey
					}
				}
			}
		}
	} else {
		// No --env flag, use original priority: profile > env vars
		manager, err := NewConfigManager()
		if err == nil {
			profile, _, err := manager.GetCurrentProfile()
			if err == nil {
				accessKey = profile.AccessKey
				secretKey = profile.SecretKey
			}
		}

		if accessKey == "" {
			accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		}
		if secretKey == "" {
			secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
		}
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
		// Capitalize the first letter of environment for display
		envDisplay := env
		if len(envDisplay) > 0 {
			envDisplay = strings.ToUpper(envDisplay[:1]) + envDisplay[1:]
		}
		spinner.StopWithSuccess(fmt.Sprintf("Successfully logged in to Megaport %s", envDisplay))
	}

	return megaportClient, nil
}

// LoginFuncWithOutput logs into the Megaport API using the current profile or environment variables.
var LoginFuncWithOutput = func(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	var accessKey, secretKey, env string

	// Environment selection priority: --env flag > profile > env var > default
	if utils.Env != "" {
		env = utils.Env
	} else {
		// Check profile environment
		manager, err := NewConfigManager()
		if err == nil {
			profile, _, err := manager.GetCurrentProfile()
			if err == nil && profile.Environment != "" {
				env = profile.Environment
			}
		}

		// Fall back to environment variable if still not set
		if env == "" {
			env = os.Getenv("MEGAPORT_ENVIRONMENT")
		}
	}

	if env == "" {
		env = "production"
	}

	// Credential selection: if --env flag is used, prefer env vars over profile
	if utils.Env != "" {
		// --env flag was explicitly set, prioritize environment variables
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

		// If env vars are empty, fall back to profile
		if accessKey == "" || secretKey == "" {
			manager, err := NewConfigManager()
			if err == nil {
				profile, _, err := manager.GetCurrentProfile()
				if err == nil {
					if accessKey == "" {
						accessKey = profile.AccessKey
					}
					if secretKey == "" {
						secretKey = profile.SecretKey
					}
				}
			}
		}
	} else {
		// No --env flag, use original priority: profile > env vars
		manager, err := NewConfigManager()
		if err == nil {
			profile, _, err := manager.GetCurrentProfile()
			if err == nil {
				accessKey = profile.AccessKey
				secretKey = profile.SecretKey
			}
		}

		if accessKey == "" {
			accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		}
		if secretKey == "" {
			secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
		}
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

	spinner := output.PrintLoggingInWithOutput(false, outputFormat)
	_, err = megaportClient.Authorize(ctx)

	if err != nil {
		spinner.Stop()
		return nil, err
	} else {
		// Capitalize the first letter of environment for display
		envDisplay := env
		if len(envDisplay) > 0 {
			envDisplay = strings.ToUpper(envDisplay[:1]) + envDisplay[1:]
		}
		spinner.StopWithSuccess(fmt.Sprintf("Successfully logged in to Megaport %s", envDisplay))
	}

	return megaportClient, nil
}
