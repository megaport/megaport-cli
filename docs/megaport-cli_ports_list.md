# list

List all ports with optional filters

## Description

List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as
port ID, name, location, speed, and status. You can optionally filter the results 
by passing additional flags such as --location-id, --port-speed, and --port-name.

Example:
```
  megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "PortName"

```

If no filtering options are provided, all ports will be listed.



## Usage

```
megaport-cli ports list [flags]
```

## Examples

```
Example:
megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "PortName"

If no filtering options are provided, all ports will be listed.
```

## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--location-id` |  | `0` | Filter ports by location ID | false |
| `--port-name` |  |  | Filter ports by port name | false |
| `--port-speed` |  | `0` | Filter ports by port speed | false |



