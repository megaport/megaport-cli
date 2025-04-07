# update

Update an existing MVE

## Description

Update an existing Megaport Virtual Edge (MVE).

This command allows you to update specific properties of an existing MVE without disrupting its service or connectivity. Updates apply immediately but may take a few minutes to fully propagate in the Megaport system.

### Optional Fields
  - `contract-term`: The new contract term in months (1, 12, 24, or 36)
  - `cost-centre`: The new cost center for billing purposes (optional)
  - `name`: The new name of the MVE (1-64 characters)

### Important Notes
  - The MVE UID cannot be changed
  - Vendor configuration cannot be changed after provisioning
  - Technical specifications (size, location) cannot be modified
  - Connectivity (VXCs) will not be affected by these changes
  - Changing the contract term may affect billing immediately

### Example Usage

```
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Edge Router West" --cost-centre "IT-Network-2023" --contract-term 24
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name": "Edge Router West", "costCentre": "IT-Network-2023", "contractTermMonths": 24}'
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json
```
### JSON Format Example
```json
{
  "name": "Edge Router West",
  "costCentre": "IT-Network-2023",
  "contractTermMonths": 24
}

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



