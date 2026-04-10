package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// AuthStatus authenticates with the Megaport API and displays the current user identity.
func AuthStatus(cmd *cobra.Command, _ []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Authentication failed: %v", noColor, err)
		return exitcodes.NewAuthError(err)
	}
	defer cancel()

	spinner := output.PrintResourceGetting("authentication status", "", noColor)

	users, err := listCompanyUsersFunc(ctx, client)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve user information: %v", noColor, err)
		return exitcodes.NewAPIError(fmt.Errorf("failed to retrieve user information: %w", err))
	}

	// Resolve profile and environment info using the same precedence as config.Login
	profileName, environment := resolveProfileInfo()
	apiEndpoint := client.BaseURL.String()

	// Find the current user from the company users list.
	// The authenticated user is typically the first active admin, but since we
	// can't determine the exact user from the token alone, we display the
	// company context with all relevant details.
	currentUser := findCurrentUser(users)

	// Derive company name from a known non-nil user to avoid panic on nil entries.
	companyName := ""
	if currentUser != nil {
		companyName = currentUser.CompanyName
	} else {
		for _, u := range users {
			if u != nil {
				companyName = u.CompanyName
				break
			}
		}
	}

	return printAuthStatus(currentUser, profileName, environment, apiEndpoint, companyName, outputFormat, noColor)
}

// findCurrentUser attempts to identify the current user from the company user list.
// It prioritizes active company admins, then any active user.
func findCurrentUser(users []*megaport.User) *megaport.User {
	if len(users) == 0 {
		return nil
	}

	// If there's only one user, that's our user
	if len(users) == 1 {
		return users[0]
	}

	// Look for active company admins first
	for _, u := range users {
		if u == nil || !u.Active {
			continue
		}
		for _, role := range u.SecurityRoles {
			if strings.EqualFold(role, "companyAdmin") {
				return u
			}
		}
	}

	// Fall back to first active user
	for _, u := range users {
		if u != nil && u.Active {
			return u
		}
	}

	return users[0]
}

// resolveProfileInfo determines the active profile name and environment using
// the same precedence rules as config.Login / resolveEnvironment:
//
// Profile: --profile flag > active profile from config > "(env vars)"
// Environment: --env flag > named profile env > active profile env > MEGAPORT_ENVIRONMENT > "production"
//
// The returned environment is always normalized to a canonical name:
// "production", "staging", or "development".
func resolveProfileInfo() (profileName, environment string) {
	profileName = "(env vars)"

	// Build environment using the same precedence as config.resolveEnvironment:
	// track it as empty until explicitly set, then normalize at the end.
	var env string

	if utils.ProfileOverride != "" {
		// --profile flag is set: read that specific named profile (same as config.Login)
		profileName = utils.ProfileOverride
		manager, err := config.NewConfigManager()
		if err == nil {
			profile, err := manager.GetProfile(utils.ProfileOverride)
			if err == nil && profile.Environment != "" {
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
			// No --env flag: read the active profile from config
			manager, err := config.NewConfigManager()
			if err == nil {
				profile, name, err := manager.GetCurrentProfile()
				if err == nil {
					profileName = name
					if profile.Environment != "" {
						env = profile.Environment
					}
				}
			}
			if env == "" {
				env = os.Getenv("MEGAPORT_ENVIRONMENT")
			}
		}
	}

	environment = normalizeEnvironment(env)
	return profileName, environment
}

// normalizeEnvironment converts environment strings to canonical names,
// matching the behavior of config.normalizeEnvironment.
func normalizeEnvironment(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return "production"
	case "staging":
		return "staging"
	case "development", "dev":
		return "development"
	default:
		return "production"
	}
}
