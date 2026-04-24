# list

List all managed accounts

## Description

List all managed accounts linked to your partner account.

This command fetches and displays a list of managed accounts with details such as account name, account reference, and company UID.

### Optional Fields
  - `account-name`: Filter managed accounts by name (partial match)
  - `account-ref`: Filter managed accounts by reference (partial match)

### Example Usage

```sh
  megaport-cli managed-account list
  megaport-cli managed-account list --account-name "Acme"
  megaport-cli managed-account list --account-ref "REF-001"
```

## Usage

```sh
megaport-cli managed-account list [flags]
```


## Parent Command

* [megaport-cli managed-account](megaport-cli_managed-account.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--account-name` |  |  | Filter managed accounts by name (partial match) | false |
| `--account-ref` |  |  | Filter managed accounts by reference (partial match) | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |

## Subcommands
* [docs](megaport-cli_managed-account_list_docs.md)

