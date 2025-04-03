# update

Update an existing MVE

## Description

Update an existing Megaport Virtual Edge (MVE).

This command allows you to update specific properties of an existing MVE without
disrupting its service or connectivity. Updates apply immediately but may take
a few minutes to fully propagate in the Megaport system.

You can provide update details in one of three ways:

1. Interactive Mode (default):
   The command will prompt you for each updatable field, showing current values
   and allowing you to make changes. Press ENTER to keep the current value.

2. Flag Mode:
   Provide only the fields you want to update as flags. Fields not specified
   will remain unchanged:
   --name, --cost-centre, --contract-term

3. JSON Mode:
   Provide a JSON string or file with the fields you want to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- `name`: The new name of the MVE (1-64 characters)
- `cost_centre`: The new cost center for billing purposes (optional)
- `contract_term_months`: The new contract term in months (1, 12, 24, or 36)

Important notes:
- The MVE UID cannot be changed
- Vendor configuration cannot be changed after provisioning
- Technical specifications (size, location) cannot be modified
- Connectivity (VXCs) will not be affected by these changes
- Changing the contract term may affect billing immediately

Example usage:

### Interactive mode (default)
```
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

```

### Flag mode
```
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Edge Router West" --cost-centre "IT-Network-2023" --contract-term 24

```

### JSON mode with string
```
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name": "Edge Router West", "costCentre": "IT-Network-2023", "contractTermMonths": 24}'

```

### JSON mode with file
```
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json

```

JSON format example (mve-update.json):
```
{
  "name": "Edge Router West",
  "costCentre": "IT-Network-2023",
  "contractTermMonths": 24
}

```

Note the JSON property names differ from flag names:
- `Flag`: --name             → JSON: "name"
- `Flag`: --cost-centre      → JSON: "costCentre"
- `Flag`: --contract-term    → JSON: "contractTermMonths"

Example successful output:
```
  MVE updated successfully:
  UID:          1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  Name:         Edge Router West (previously "Edge Router")  
  Cost Centre:  IT-Network-2023 (previously "IT-Network")
  Term:         24 months (previously 12 months)

```



## Usage

```
megaport-cli mve update [mveUID] [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--contract-term` |  | `0` | New contract term in months (1, 12, 24, or 36) | false |
| `--cost-centre` |  |  | New cost centre | false |
| `--interactive` | `-i` | `true` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing MVE update configuration | false |
| `--json-file` |  |  | Path to JSON file containing MVE update configuration | false |
| `--name` |  |  | New MVE name | false |



