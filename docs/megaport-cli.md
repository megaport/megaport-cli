# megaport-cli

A CLI tool to interact with the Megaport API

## Description

A CLI tool to interact with the Megaport API.

This CLI supports the following features:
- `Locations`: List and manage locations.
- `Ports`: List all ports and get details for a specific port.
- `MCRs`: Get details for Megaport Cloud Routers.
- `MVEs`: Get details for Megaport Virtual Edge devices.
- `VXCs`: Get details for Virtual Cross Connects.
- `Partner Ports`: List and filter partner ports based on product name, connect type, company name, location ID, and diversity zone.



## Usage

```
megaport-cli [flags]
```







## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --env | -e | production | Environment to use (production, staging, development) | false |
| --output | -o | table | Output format (table, json, csv, xml) | false |
| --env | -e | production | Environment to use (production, staging, development) | false |
| --output | -o | table | Output format (table, json, csv, xml) | false |


## Subcommands

* [completion](megaport-cli_completion.md)
* [generate-docs](megaport-cli_generate-docs.md)
* [locations](megaport-cli_locations.md)
* [mcr](megaport-cli_mcr.md)
* [mve](megaport-cli_mve.md)
* [partners](megaport-cli_partners.md)
* [ports](megaport-cli_ports.md)
* [servicekeys](megaport-cli_servicekeys.md)
* [version](megaport-cli_version.md)
* [vxc](megaport-cli_vxc.md)

