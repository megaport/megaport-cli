# update

Update a port's details

## Description

Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
The command will prompt you for each updatable field, showing current values
and allowing you to make changes. Press ENTER to keep the current value.

2. Flag Mode:
Provide only the fields you want to update as flags:
--name, --marketplace-visibility, --cost-centre, --term

3. JSON Mode:
Provide a JSON string or file with the fields you want to update:
--json <json-string> or --json-file <path>

Fields that can be updated:
- name: The new name of the port (1-64 characters)
- marketplace_visibility: Whether the port should be visible in the marketplace (true or false)
- cost_centre: The cost center for billing purposes (optional)
- term: The new contract term in months (1, 12, 24, or 36)

Important notes:
- The port UID cannot be changed
- Technical specifications (speed, location) cannot be modified
- Connectivity (VXCs) will not be affected by these changes
- Changing the contract term may affect billing immediately

Example usage:

### Interactive mode
```
megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --interactive

```
### Flag mode
```
megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Main Data Center Port" --marketplace-visibility false

```
### JSON mode
```
megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name":"Main Data Center Port","marketplaceVisibility":false}'
megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./update-port-config.json

JSON format example (update-port-config.json):
{
"name": "Main Data Center Port",
"marketplaceVisibility": false,
"costCentre": "IT-Network-2023",
"term": 24
}

Note the JSON property names differ from flag names:
- Flag: --name                      → JSON: "name"
- Flag: --marketplace-visibility    → JSON: "marketplaceVisibility"
- Flag: --cost-centre               → JSON: "costCentre"
- Flag: --term                      → JSON: "term"

Example successful output:
Port updated successfully - UID: 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

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



