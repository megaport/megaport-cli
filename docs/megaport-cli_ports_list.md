# list

List all ports with optional filters

## Description

List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as port ID, name, location, speed, and status. By default, only active ports are shown. You can also filter by resource tags.

### Optional Fields
  - `include-inactive`: Include ports in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states
  - `location-id`: Filter ports by location ID
  - `port-name`: Filter ports by port name
  - `port-speed`: Filter ports by port speed
  - `tag`: Filter by resource tag (format: key=value or key; repeatable, AND logic)

### Example Usage

```sh
  megaport-cli ports list
  megaport-cli ports list --location-id 1
  megaport-cli ports list --port-speed 10000
  megaport-cli ports list --port-name "Data Center Primary"
  megaport-cli ports list --include-inactive
  megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "Data Center Primary"
  megaport-cli ports list --tag env=prod
  megaport-cli ports list --tag env=prod --tag team=network
```

## Usage

```sh
megaport-cli ports list [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive ports in the list | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--location-id` |  | `0` | Filter ports by location ID | false |
| `--port-name` |  |  | Filter ports by port name | false |
| `--port-speed` |  | `0` | Filter ports by port speed | false |
| `--tag` |  | `[]` | Filter by resource tag (format: key=value or key; repeatable, AND logic) | false |

