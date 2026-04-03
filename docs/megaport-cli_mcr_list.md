# list

List all MCRs with optional filters

## Description

List all MCRs available in the Megaport API.

This command fetches and displays a list of MCRs with details such as MCR ID, name, location, speed, and status. By default, only active MCRs are shown.

### Optional Fields
  - `include-inactive`: Include MCRs in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states
  - `location-id`: Filter MCRs by location ID
  - `name`: Filter MCRs by name
  - `port-speed`: Filter MCRs by port speed

### Example Usage

```sh
  megaport-cli mcr list
  megaport-cli mcr list --location-id 1
  megaport-cli mcr list --port-speed 10000
  megaport-cli mcr list --name "My MCR"
  megaport-cli mcr list --include-inactive
  megaport-cli mcr list --location-id 1 --port-speed 10000 --name "My MCR"
```

## Usage

```sh
megaport-cli mcr list [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive MCRs in the list | false |
| `--location-id` |  | `0` | Filter MCRs by location ID | false |
| `--name` |  |  | Filter MCRs by name | false |
| `--port-speed` |  | `0` | Filter MCRs by port speed | false |

## Subcommands
* [docs](megaport-cli_mcr_list_docs.md)

