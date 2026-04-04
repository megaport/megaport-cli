//go:build !js && !wasm
// +build !js,!wasm

package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// resolveEnvironment determines the target API environment using the following
// priority: --env flag > profile config > MEGAPORT_ENVIRONMENT env var > default (production).
// When --profile is set, the named profile's environment is used as a base,
// but --env flag still overrides it.
// If requireProfile is true, errors are returned when --profile is set but the
// profile cannot be loaded; otherwise profile errors are silently ignored.
func resolveEnvironment(requireProfile bool) (string, error) {
	var env string

	if utils.ProfileOverride != "" {
		manager, err := NewConfigManager()
		if err != nil {
			if requireProfile {
				return "", fmt.Errorf("failed to load config for profile %q: %w", utils.ProfileOverride, err)
			}
		} else {
			profile, err := manager.GetProfile(utils.ProfileOverride)
			if err != nil {
				if requireProfile {
					return "", fmt.Errorf("profile %q not found. Use 'megaport config list-profiles' to see available profiles", utils.ProfileOverride)
				}
			} else if profile.Environment != "" {
				env = profile.Environment
			}
		}
		// --env flag overrides the profile's environment
		if utils.Env != "" {
			env = utils.Env
		}
		// Fall back to env var if still not set
		if env == "" {
			env = os.Getenv("MEGAPORT_ENVIRONMENT")
		}
	} else {
		if utils.Env != "" {
			env = utils.Env
		} else {
			manager, err := NewConfigManager()
			if err == nil {
				profile, _, err := manager.GetCurrentProfile()
				if err == nil && profile.Environment != "" {
					env = profile.Environment
				}
			}
			if env == "" {
				env = os.Getenv("MEGAPORT_ENVIRONMENT")
			}
		}
	}

	return env, nil
}

// environmentOption returns the megaport.ClientOpt for the given environment string.
func environmentOption(env string) megaport.ClientOpt {
	switch env {
	case "production":
		return megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		return megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		return megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		return megaport.WithEnvironment(megaport.EnvironmentProduction)
	}
}

func Login(ctx context.Context) (*megaport.Client, error) {
	return LoginFunc(ctx)
}

func LoginWithOutput(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	return LoginFuncWithOutput(ctx, outputFormat)
}

// LoginFunc logs into the Megaport API using the current profile or environment variables.
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	return LoginFuncWithOutput(ctx, "")
}

// LoginFuncWithOutput logs into the Megaport API using the current profile or environment variables.
var LoginFuncWithOutput = func(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	var accessKey, secretKey string

	env, err := resolveEnvironment(false)
	if err != nil {
		return nil, err
	}

	// If --profile flag is set, use that specific profile directly
	if utils.ProfileOverride != "" {
		manager, err := NewConfigManager()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		profile, err := manager.GetProfile(utils.ProfileOverride)
		if err != nil {
			return nil, fmt.Errorf("profile %q not found. Use 'megaport config list-profiles' to see available profiles", utils.ProfileOverride)
		}
		accessKey = profile.AccessKey
		secretKey = profile.SecretKey
	} else {
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
	}

	if accessKey == "" {
		return nil, fmt.Errorf("megaport API access key not provided. Configure an active profile or set MEGAPORT_ACCESS_KEY environment variable")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("megaport API secret key not provided. Configure an active profile or set MEGAPORT_SECRET_KEY environment variable")
	}

	envOpt := environmentOption(env)
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

// NewUnauthenticatedClientFunc creates a Megaport API client without authentication.
// Used for public API endpoints (e.g., locations) that don't require credentials.
var NewUnauthenticatedClientFunc = func() (*megaport.Client, error) {
	env, err := resolveEnvironment(true)
	if err != nil {
		return nil, err
	}

	envOpt := environmentOption(env)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return megaport.New(httpClient, envOpt)
}

// NewUnauthenticatedClient creates an unauthenticated Megaport API client.
func NewUnauthenticatedClient() (*megaport.Client, error) {
	return NewUnauthenticatedClientFunc()
}
