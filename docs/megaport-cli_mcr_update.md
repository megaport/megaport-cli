# update

Update an existing MCR

## Description

Update an existing Megaport Cloud Router (MCR).

This command allows you to update the details of an existing MCR.

### Optional Fields
  - `cost-centre`: The new cost centre for the MCR
  - `marketplace-visibility`: Whether the MCR is visible in the marketplace (true/false)
  - `name`: The new name of the MCR (1-64 characters)

### Important Notes
  - The MCR UID cannot be changed
  - Only specified fields will be updated; unspecified fields will remain unchanged
  - Ensure the JSON file is correctly formatted

### Example Usage

```sh
  megaport-cli mcr update [mcrUID] --interactive
  megaport-cli mcr update [mcrUID] --name "Updated MCR" --marketplace-visibility true --cost-centre "Finance"
  megaport-cli mcr update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true,"costCentre":"Finance"}'
  megaport-cli mcr update [mcrUID] --json-file ./update-mcr-config.json
```
### JSON Format Example
```json
{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "contractTermMonths": 24
}

```

## Usage

```sh
megaport-cli mcr update [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | The new cost centre for the MCR | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--marketplace-visibility` |  | `false` | Whether the MCR is visible in the marketplace (true/false) | false |
| `--name` |  |  | The new name of the MCR (1-64 characters) | false |

