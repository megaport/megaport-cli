# update

Update a port's details

## Description

Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.

### Optional Fields
  - `cost-centre`: The cost centre for billing purposes
  - `marketplace-visibility`: Whether the port should be visible in the marketplace (true or false)
  - `name`: The new name of the port (1-64 characters)
  - `term`: The new contract term in months (1, 12, 24, or 36)

### Important Notes
  - The port UID cannot be changed
  - Technical specifications (speed, location) cannot be modified
  - Connectivity (VXCs) will not be affected by these changes
  - Changing the contract term may affect billing immediately

### Example Usage

```
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --interactive
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Main Data Center Port" --marketplace-visibility false
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name":"Main Data Center Port","marketplaceVisibility":false}'
  update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./update-port-config.json
```
### JSON Format Example
```json
{
  "name": "Main Data Center Port",
  "marketplaceVisibility": false,
  "costCentre": "IT-Network-2023",
  "term": 24
}

```


## Usage

```
megaport-cli ports update [portUID] [flags]
```



## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing port configuration | false |
| `--json-file` |  |  | Path to JSON file containing port configuration | false |
| `--marketplace-visibility` |  | `false` | Whether the port is visible in marketplace | false |
| `--name` |  |  | New port name | false |
| `--term` |  | `0` | New contract term in months (1, 12, 24, or 36) | false |



