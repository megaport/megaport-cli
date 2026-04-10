package auth

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the auth commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	authCmd := cmdbuilder.NewCommand("auth", "Manage authentication and view current identity").
		WithLongDesc("Manage authentication and view the currently authenticated identity.\n\n" +
			"Use the status subcommand to verify your credentials and see which user, " +
			"company, environment, and profile you are currently operating as.").
		WithExample("megaport-cli auth status").
		WithExample("megaport-cli auth status --output json").
		WithRootCmd(rootCmd).
		Build()

	statusCmd := cmdbuilder.NewCommand("status", "Display current authentication status and identity").
		WithOutputFormatRunFunc(AuthStatus).
		WithLongDesc("Verify your credentials and display the current account context.\n\n" +
			"This command authenticates with the Megaport API using your active profile or " +
			"environment variables, then retrieves your company user details. It shows the " +
			"company, environment, active profile, and API endpoint.\n\n" +
			"Note: the displayed user is inferred from the company user list (preferring the " +
			"primary admin). For companies with multiple admins, it may not reflect the exact " +
			"user who owns the API credentials.\n\n" +
			"Use this to confirm which account and environment you are operating against before " +
			"making infrastructure changes.").
		WithExample("megaport-cli auth status").
		WithExample("megaport-cli auth status --output json").
		WithExample("megaport-cli auth status --output json --query 'email'").
		WithRootCmd(rootCmd).
		Build()

	// whoami is a top-level convenience alias for "auth status"
	whoamiCmd := cmdbuilder.NewCommand("whoami", "Display current authenticated identity").
		WithOutputFormatRunFunc(AuthStatus).
		WithLongDesc("Display the currently authenticated identity (alias for 'auth status').\n\n" +
			"Verifies your credentials and shows your user details, active profile, " +
			"environment, and API endpoint.").
		WithExample("megaport-cli whoami").
		WithExample("megaport-cli whoami --output json").
		WithRootCmd(rootCmd).
		Build()

	authCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(whoamiCmd)
}
