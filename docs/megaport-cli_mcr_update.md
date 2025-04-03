# update

Update an existing MCR

## Description

Update an existing Megaport Cloud Router (MCR).

This command allows you to update the details of an existing MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each field you want to update, showing the current value and allowing you to modify it.

2. Flag Mode:
   Provide fields as flags:
   --name, --cost-centre, --marketplace-visibility, --term
   Only specified flags will be updated; unspecified fields will remain unchanged.

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>
   Only fields present in the JSON will be updated; unspecified fields will remain unchanged.

Fields that can be updated:
- `name`: The new name of the MCR (1-64 characters).
- `cost_centre`: The new cost center for the MCR.
- `marketplace_visibility`: Whether the MCR is visible in the marketplace (true/false).
- `term`: The new contract term in months (1, 12, 24, or 36).

Example usage:

### Interactive mode
```
  megaport-cli mcr update [mcrUID] --interactive

```

### Flag mode
```
  megaport-cli mcr update [mcrUID] --name "Updated MCR" --marketplace-visibility true --cost-centre "Finance"

```

### JSON mode
```
  megaport-cli mcr update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true,"costCentre":"Finance"}'
  megaport-cli mcr update [mcrUID] --json-file ./update-mcr-config.json

```

JSON format example (update-mcr-config.json):
```
{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "term": 24
}

```

Notes:
- The MCR UID cannot be changed.
- Only specified fields will be updated; unspecified fields will remain unchanged.
- Ensure the JSON file is correctly formatted.



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



