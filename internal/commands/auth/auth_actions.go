package auth

import (
	"fmt"
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

	// Resolve profile and environment info
	profileName, environment := resolveProfileInfo()
	apiEndpoint := client.BaseURL.String()

	// Find the current user from the company users list.
	// The authenticated user is typically the first active admin, but since we
	// can't determine the exact user from the token alone, we display the
	// company context with all relevant details.
	var currentUser *megaport.User
	companyName := ""

	if len(users) > 0 {
		// All users share the same company, grab it from the first
		companyName = users[0].CompanyName
		// Try to find the most likely current user (first active admin)
		currentUser = findCurrentUser(users)
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

// resolveProfileInfo reads the active profile name and environment from config.
func resolveProfileInfo() (profileName, environment string) {
	profileName = "(env vars)"
	environment = "production"

	// Check if --profile override is in use
	if utils.ProfileOverride != "" {
		profileName = utils.ProfileOverride
	}

	// Try to load profile config for environment and profile name
	manager, err := config.NewConfigManager()
	if err == nil {
		profile, name, err := manager.GetCurrentProfile()
		if err == nil {
			if utils.ProfileOverride == "" {
				profileName = name
			}
			if profile.Environment != "" {
				environment = profile.Environment
			}
		}
	}

	// --env flag always overrides regardless of config state
	if utils.Env != "" {
		environment = utils.Env
	}

	return profileName, environment
}
