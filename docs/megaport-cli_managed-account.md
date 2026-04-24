# managed-account

Manage partner managed accounts in the Megaport API

## Description

Manage partner managed accounts in the Megaport API.

This command groups all operations related to Megaport managed accounts. Managed accounts allow Megaport Partners to create and manage sub-accounts (companies) linked to their partner account.

### Important Notes
  - Managed accounts are a partner-only feature
  - Each managed account represents a sub-company under the partner's umbrella

### Example Usage

```sh
  megaport-cli managed-account list
  megaport-cli managed-account get [companyUID] [accountName]
  megaport-cli managed-account create
  megaport-cli managed-account update [companyUID]
```

## Usage

```sh
megaport-cli managed-account [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [create](megaport-cli_managed-account_create.md)
* [docs](megaport-cli_managed-account_docs.md)
* [get](megaport-cli_managed-account_get.md)
* [list](megaport-cli_managed-account_list.md)
* [update](megaport-cli_managed-account_update.md)

