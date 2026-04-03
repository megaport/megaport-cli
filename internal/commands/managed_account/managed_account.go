package managed_account

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the managed-account commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create managed-account parent command
	managedAccountCmd := cmdbuilder.NewCommand("managed-account", "Manage partner managed accounts in the Megaport API").
		WithLongDesc("Manage partner managed accounts in the Megaport API.\n\nThis command groups all operations related to Megaport managed accounts. Managed accounts allow Megaport Partners to create and manage sub-accounts (companies) linked to their partner account.").
		WithExample("megaport-cli managed-account list").
		WithExample("megaport-cli managed-account get [companyUID] [accountName]").
		WithExample("megaport-cli managed-account create").
		WithExample("megaport-cli managed-account update [companyUID]").
		WithImportantNote("Managed accounts are a partner-only feature").
		WithImportantNote("Each managed account represents a sub-company under the partner's umbrella").
		WithRootCmd(rootCmd).
		Build()

	// Create list managed accounts command
	listCmd := cmdbuilder.NewCommand("list", "List all managed accounts").
		WithOutputFormatRunFunc(ListManagedAccounts).
		WithLongDesc("List all managed accounts linked to your partner account.\n\nThis command fetches and displays a list of managed accounts with details such as account name, account reference, and company UID.").
		WithManagedAccountFilterFlags().
		WithOptionalFlag("account-name", "Filter managed accounts by name (partial match)").
		WithOptionalFlag("account-ref", "Filter managed accounts by reference (partial match)").
		WithExample("megaport-cli managed-account list").
		WithExample("megaport-cli managed-account list --account-name \"Acme\"").
		WithExample("megaport-cli managed-account list --account-ref \"REF-001\"").
		WithIntFlag("limit", 0, "Maximum number of results to display (0 = unlimited)").
		WithRootCmd(rootCmd).
		Build()

	// Create get managed account command
	getCmd := cmdbuilder.NewCommand("get", "Get details for a single managed account").
		WithArgs(cobra.ExactArgs(2)).
		WithOutputFormatRunFunc(GetManagedAccount).
		WithLongDesc("Get details for a single managed account.\n\nThis command retrieves and displays detailed information for a single managed account. You must provide the company UID and account name.").
		WithExample("megaport-cli managed-account get [companyUID] [accountName]").
		WithImportantNote("The first argument is the company UID and the second is the account name").
		WithRootCmd(rootCmd).
		Build()

	// Create create managed account command
	createCmd := cmdbuilder.NewCommand("create", "Create a new managed account").
		WithColorAwareRunFunc(CreateManagedAccount).
		WithManagedAccountCreateFlags().
		WithStandardInputFlags().
		WithLongDesc("Create a new managed account through the Megaport API.\n\nThis command allows you to create a new managed account (sub-company) under your partner account.").
		WithDocumentedRequiredFlag("account-name", "The name of the managed account").
		WithDocumentedRequiredFlag("account-ref", "The reference ID for the managed account").
		WithExample("megaport-cli managed-account create --interactive").
		WithExample("megaport-cli managed-account create --account-name \"Acme Corp\" --account-ref \"REF-001\"").
		WithExample("megaport-cli managed-account create --json '{\"accountName\":\"Acme Corp\",\"accountRef\":\"REF-001\"}'").
		WithExample("megaport-cli managed-account create --json-file ./account-config.json").
		WithJSONExample(`{
  "accountName": "Acme Corp",
  "accountRef": "REF-001"
}`).
		WithImportantNote("Required flags (account-name, account-ref) can be skipped when using --interactive, --json, or --json-file").
		WithRootCmd(rootCmd).
		WithConditionalRequirements("account-name", "account-ref").
		Build()

	// Create update managed account command
	updateCmd := cmdbuilder.NewCommand("update", "Update an existing managed account").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateManagedAccount).
		WithStandardInputFlags().
		WithManagedAccountUpdateFlags().
		WithLongDesc("Update an existing managed account.\n\nThis command allows you to update the details of an existing managed account.").
		WithExample("megaport-cli managed-account update [companyUID] --interactive").
		WithExample("megaport-cli managed-account update [companyUID] --account-name \"New Name\"").
		WithExample("megaport-cli managed-account update [companyUID] --json '{\"accountName\":\"New Name\",\"accountRef\":\"REF-002\"}'").
		WithExample("megaport-cli managed-account update [companyUID] --json-file ./update-config.json").
		WithJSONExample(`{
  "accountName": "Updated Corp",
  "accountRef": "REF-002"
}`).
		WithImportantNote("The company UID cannot be changed").
		WithImportantNote("Only specified fields will be updated; unspecified fields will remain unchanged").
		WithRootCmd(rootCmd).
		Build()

	// Note: No delete command — the ManagedAccountService SDK does not expose a delete operation.

	// Add commands to their parents
	managedAccountCmd.AddCommand(
		listCmd,
		getCmd,
		createCmd,
		updateCmd,
	)
	rootCmd.AddCommand(managedAccountCmd)
}
