//go:build !js || !wasm
// +build !js !wasm

package config

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

func AddCommandsTo(rootCmd *cobra.Command) {
	configCmd := cmdbuilder.NewCommand("config", "Manage configuration settings").
		WithLongDesc("Manage configuration settings for Megaport CLI.\n\n" +
			"The config command allows you to manage persistent configuration settings for the CLI, " +
			"including authentication profiles with environment settings. " +
			"Profiles store your API credentials and environment settings " +
			"for streamlined operations across multiple Megaport environments.\n\n" +
			"Configuration is stored locally in ~/.megaport/config.json and persists across CLI sessions.\n\n" +
			"Configuration Precedence:\n" +
			"1. Command-line flags (highest precedence)\n" +
			"2. Environment variables (MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, etc.)\n" +
			"3. Active profile in config file\n" +
			"4. Default settings in config file (lowest precedence)").
		WithExample("megaport-cli config create-profile production --access-key xxx --secret-key xxx --environment production").
		WithExample("megaport-cli config use-profile production").
		WithImportantNote("Configuration contains sensitive credentials - ensure ~/.megaport directory has appropriate permissions").
		WithImportantNote("Environment variables (MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY) take precedence over stored profiles").
		WithRootCmd(rootCmd).
		Build()

	createProfileCmd := cmdbuilder.NewCommand("create-profile", "Create a new credential profile").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(CreateProfile).
		WithLongDesc("Create a new profile with Megaport API credentials and environment settings.\n\n"+
			"Profiles store your Megaport API access and secret keys along with environment settings for secure reuse. "+
			"The profile name is case-sensitive and must be unique.\n\n"+
			"Credentials are stored in ~/.megaport/config.json with secure file permissions.").
		WithFlag("access-key", "", "Megaport API access key from the Megaport Portal").
		WithRequiredFlag("access-key", "Megaport API access key from the Megaport Portal").
		WithFlag("secret-key", "", "Megaport API secret key from the Megaport Portal").
		WithRequiredFlag("secret-key", "Megaport API secret key from the Megaport Portal").
		WithFlag("environment", "production", "Target API environment: 'production', 'staging', or 'development'").
		WithFlag("description", "", "Optional description for this profile").
		WithExample("megaport-cli config create-profile production --access-key xxx --secret-key xxx --environment production --description \"Production credentials\"").
		WithExample("megaport-cli config create-profile staging --access-key yyy --secret-key yyy --environment staging").
		WithImportantNote("API credentials are stored with 0600 permissions (readable only by the current user)").
		WithRootCmd(rootCmd).
		Build()

	updateProfileCmd := cmdbuilder.NewCommand("update-profile", "Update an existing profile").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateProfile).
		WithLongDesc("Update an existing profile with new credentials or settings.").
		WithFlag("access-key", "", "Megaport API access key").
		WithFlag("secret-key", "", "Megaport API secret key").
		WithFlag("environment", "", "Target API environment: 'production', 'staging', or 'development'").
		WithFlag("description", "", "Profile description (use empty string to clear)").
		WithExample("megaport-cli config update-profile myprofile --access-key xxx --secret-key xxx").
		WithExample("megaport-cli config update-profile myprofile --environment staging").
		WithImportantNote("Keep your Megaport API credentials secure; they provide full account access").
		WithRootCmd(rootCmd).
		Build()

	deleteProfileCmd := cmdbuilder.NewCommand("delete-profile", "Delete a profile").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(DeleteProfile).
		WithLongDesc("Delete a profile from your configuration.").
		WithExample("megaport-cli config delete-profile myprofile").
		WithRootCmd(rootCmd).
		Build()

	listProfilesCmd := cmdbuilder.NewCommand("list-profiles", "List all profiles").
		WithOutputFormatRunFunc(ListProfiles).
		WithLongDesc("List all profiles with their associated access keys and environments.").
		WithExample("megaport-cli config list-profiles").
		WithRootCmd(rootCmd).
		Build()

	useProfileCmd := cmdbuilder.NewCommand("use-profile", "Switch to a profile").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UseProfile).
		WithLongDesc("Set the active profile for CLI operations.").
		WithExample("megaport-cli config use-profile myprofile").
		WithRootCmd(rootCmd).
		Build()

	setDefaultCmd := cmdbuilder.NewCommand("set-default", "Set a default value").
		WithArgs(cobra.ExactArgs(2)).
		WithColorAwareRunFunc(SetDefault).
		WithLongDesc("Set a default value in the configuration.").
		WithExample("megaport-cli config set-default output json").
		WithExample("megaport-cli config set-default no-color true").
		WithRootCmd(rootCmd).
		Build()

	getDefaultCmd := cmdbuilder.NewCommand("get-default", "Get a default value").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(GetDefault).
		WithLongDesc("Get a default value from the configuration.").
		WithExample("megaport-cli config get-default output").
		WithRootCmd(rootCmd).
		Build()

	exportCmd := cmdbuilder.NewCommand("export", "Export configuration").
		WithLongDesc("Export configuration to a file (excluding sensitive information).\n\n"+
			"The export function writes your configuration to a JSON file with sensitive "+
			"information like access keys and secret keys REDACTED for security purposes. "+
			"This means you CANNOT directly import an export file to restore credentials.\n\n"+
			"Exports are useful for:\n"+
			"- Backing up profile settings and defaults (without credentials)\n"+
			"- Sharing configuration templates with teammates\n"+
			"- Transferring settings between environments\n\n"+
			"To use an exported file on another system, you must manually edit the file "+
			"to replace [REDACTED] values with actual credentials before importing.").
		WithColorAwareRunFunc(ExportConfig).
		WithFlag("file", "", "File to export to").
		WithExample("megaport-cli config export --file myconfig.json").
		WithRootCmd(rootCmd).
		Build()

	importCmd := cmdbuilder.NewCommand("import", "Import configuration").
		WithLongDesc("Import configuration from a file.\n\n"+
			"Import allows you to load profiles and default settings from a JSON file. "+
			"The import file must follow the structure of an export file, with valid credentials "+
			"in place of any [REDACTED] values.\n\n"+
			"IMPORTANT: Importing merges with existing configuration. It will:\n"+
			"- Add new profiles that don't exist\n"+
			"- Update existing profiles with the same name\n"+
			"- Add or update default settings\n"+
			"- Set the active profile if specified in the import file\n\n"+
			"Version compatibility: Import supports config file versions up to the current version.").
		WithColorAwareRunFunc(ImportConfig).
		WithFlag("file", "", "File to import from").
		WithRequiredFlag("file", "File to import from").
		WithExample("megaport-cli config import --file myconfig.json").
		WithImportantNote("Credentials marked as [REDACTED] in export files must be replaced with actual values before import").
		WithRootCmd(rootCmd).
		Build()

	viewCmd := cmdbuilder.NewCommand("view", "Display current configuration").
		WithLongDesc("Display the current active configuration settings for the Megaport CLI.\n\n" +
			"This command shows your active profile and default settings. " +
			"Sensitive information like secret keys is partially masked for security. " +
			"Use this command to verify your current working configuration before executing commands.").
		WithColorAwareRunFunc(ViewConfig).
		WithExample("megaport-cli config view").
		WithRootCmd(rootCmd).
		Build()

	removeDefaultCmd := cmdbuilder.NewCommand("remove-default", "Remove a default value").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(RemoveDefault).
		WithLongDesc("Remove a default value from the configuration.").
		WithExample("megaport-cli config remove-default output").
		WithRootCmd(rootCmd).
		Build()

	clearDefaultsCmd := cmdbuilder.NewCommand("clear-defaults", "Clear all default settings").
		WithColorAwareRunFunc(ClearDefaults).
		WithLongDesc("Remove all default values from the configuration.").
		WithExample("megaport-cli config clear-defaults").
		WithRootCmd(rootCmd).
		Build()

	configCmd.AddCommand(
		createProfileCmd,
		updateProfileCmd,
		deleteProfileCmd,
		listProfilesCmd,
		useProfileCmd,
		setDefaultCmd,
		getDefaultCmd,
		removeDefaultCmd,
		clearDefaultsCmd,
		exportCmd,
		importCmd,
		viewCmd,
	)

	rootCmd.AddCommand(configCmd)
}
