# update

Update an existing managed account

## Description

Update an existing managed account.

This command allows you to update the details of an existing managed account.

### Optional Fields
  - `account-name`: The new name of the managed account
  - `account-ref`: The new reference ID for the managed account

### Important Notes
  - The company UID cannot be changed
  - Only specified fields will be updated; unspecified fields will remain unchanged

### Example Usage

```sh
  megaport-cli managed-account update [companyUID] --interactive
  megaport-cli managed-account update [companyUID] --account-name "New Name"
  megaport-cli managed-account update [companyUID] --json '{"accountName":"New Name","accountRef":"REF-002"}'
  megaport-cli managed-account update [companyUID] --json-file ./update-config.json
```
### JSON Format Example
```json
{
  "accountName": "Updated Corp",
  "accountRef": "REF-002"
}

```

## Usage

```sh
megaport-cli managed-account update [flags]
```


## Parent Command

* [megaport-cli managed-account](megaport-cli_managed-account.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--account-name` |  |  | The new name of the managed account | false |
| `--account-ref` |  |  | The new reference ID for the managed account | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |

## Subcommands
* [docs](megaport-cli_managed-account_update_docs.md)

