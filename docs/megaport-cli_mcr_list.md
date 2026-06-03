# list

List all MCRs with optional filters

## Description

List all MCRs available in the Megaport API.

This command fetches and displays a list of MCRs with details such as MCR ID, name, location, speed, and status. By default, only active MCRs are shown. You can also filter by resource tags.

### Optional Fields
  - `include-inactive`: Include MCRs in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states
  - `location-id`: Filter MCRs by location ID
  - `name`: Filter MCRs by name
  - `port-speed`: Filter MCRs by port speed
  - `tag`: Filter by resource tag (format: key=value or key; repeatable, AND logic)

### Example Usage

```sh
  megaport-cli mcr list
  megaport-cli mcr list --location-id 1
  megaport-cli mcr list --port-speed 10000
  megaport-cli mcr list --name "My MCR"
  megaport-cli mcr list --include-inactive
  megaport-cli mcr list --location-id 1 --port-speed 10000 --name "My MCR"
  megaport-cli mcr list --tag env=prod
  megaport-cli mcr list --tag env=prod --tag team=network
```

## Usage

```sh
megaport-cli mcr list [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive MCRs in the list | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--location-id` |  | `0` | Filter MCRs by location ID | false |
| `--name` |  |  | Filter MCRs by name | false |
| `--port-speed` |  | `0` | Filter MCRs by port speed | false |
| `--tag` |  | `[]` | Filter by resource tag (format: key=value or key; repeatable, AND logic) | false |

## Subcommands
* [docs](megaport-cli_mcr_list_docs.md)

