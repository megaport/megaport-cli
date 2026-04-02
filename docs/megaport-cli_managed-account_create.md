# create

Create a new managed account

## Description

Create a new managed account through the Megaport API.

This command allows you to create a new managed account (sub-company) under your partner account.

### Required Fields
  - `account-name`: The name of the managed account
  - `account-ref`: The reference ID for the managed account

### Important Notes
  - Required flags (account-name, account-ref) can be skipped when using --interactive, --json, or --json-file

### Example Usage

```sh
  megaport-cli managed-account create --interactive
  megaport-cli managed-account create --account-name "Acme Corp" --account-ref "REF-001"
  megaport-cli managed-account create --json '{"accountName":"Acme Corp","accountRef":"REF-001"}'
  megaport-cli managed-account create --json-file ./account-config.json
```
### JSON Format Example
```json
{
  "accountName": "Acme Corp",
  "accountRef": "REF-001"
}

```

## Usage

```sh
megaport-cli managed-account create [flags]
```


## Parent Command

* [megaport-cli managed-account](megaport-cli_managed-account.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--account-name` |  |  | The name of the managed account | true |
| `--account-ref` |  |  | The reference ID for the managed account | true |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |

## Subcommands
* [docs](megaport-cli_managed-account_create_docs.md)

