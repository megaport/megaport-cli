# update

Update an existing MCR

## Description

Update an existing Megaport Cloud Router (MCR).

This command allows you to update the details of an existing MCR.

### Optional Fields
  - `cost-centre`: The new cost centre for the MCR
  - `marketplace-visibility`: Whether the MCR is visible in the marketplace (true/false)
  - `name`: The new name of the MCR (1-64 characters)
  - `term`: The new contract term in months (1, 12, 24, or 36)

### Important Notes
  - The MCR UID cannot be changed
  - Only specified fields will be updated; unspecified fields will remain unchanged
  - Ensure the JSON file is correctly formatted

### Example Usage

```
  update [mcrUID] --interactive
  update [mcrUID] --name "Updated MCR" --marketplace-visibility true --cost-centre "Finance"
  update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true,"costCentre":"Finance"}'
  update [mcrUID] --json-file ./update-mcr-config.json
```
### JSON Format Example
```json
{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "term": 24
}

```


## Usage

```
megaport-cli mcr update [mcrUID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing MCR configuration | false |
| `--json-file` |  |  | Path to JSON file containing MCR configuration | false |
| `--marketplace-visibility` |  | `false` | Whether the MCR is visible in marketplace | false |
| `--name` |  |  | New MCR name | false |
| `--term` |  | `0` | New contract term in months (1, 12, 24, or 36) | false |



