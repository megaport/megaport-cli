# list

List all ports with optional filters

## Description

List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as port ID, name, location, speed, and status.

### Optional Fields
  - `available`: Show only available ports
  - `lag-only`: Show only LAG ports
  - `location-id`: Filter ports by location ID
  - `port-name`: Filter ports by port name
  - `port-speed`: Filter ports by port speed

### Example Usage

```sh
  megaport-cli ports list
  megaport-cli ports list --location-id 1
  megaport-cli ports list --port-speed 10000
  megaport-cli ports list --port-name "Data Center Primary"
  megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "Data Center Primary"
```

## Usage

```sh
megaport-cli ports list [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--location-id` |  | `0` | Filter ports by location ID | false |
| `--port-name` |  |  | Filter ports by port name | false |
| `--port-speed` |  | `0` | Filter ports by port speed | false |

