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
  - At least one update flag must be provided when not using --interactive, --json, or --json-file


## Usage

```sh
megaport-cli ports update [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--marketplace-visibility` |  | `false` | Whether the port is visible in marketplace | false |
| `--name` |  |  | New port name | false |
| `--term` |  | `0` | New contract term in months (1, 12, 24, or 36) | false |

## Subcommands
* [docs](megaport-cli_ports_update_docs.md)

