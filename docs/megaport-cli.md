# megaport-cli

A CLI tool to interact with the Megaport API

## Description

Megaport CLI provides a command line interface to interact with the Megaport API.

The CLI allows you to manage Megaport resources such as ports, VXCs, MCRs, MVEs, service keys, and more.

### Optional Fields
  - `--env`: Environment to use (production, staging, development)
  - `--help`: Show help for any command
  - `--no-color`: Disable colored output
  - `--output`: Output format (json, yaml, table, csv, xml)

### Important Notes
  - Use the --help flag with any command to see specific usage information
  - Authentication is handled via the MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY environment variables
  - By default, the CLI connects to the Megaport production environment
  - Set the MEGAPORT_ENDPOINT environment variable to connect to a different environment

### Example Usage

```sh
  megaport-cli ports list
  megaport-cli vxc buy --interactive
  megaport-cli mcr get [mcrUID]
  megaport-cli locations list
```

## Usage

```sh
megaport-cli [flags]
```




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--env` |  |  | Environment to use (prod, dev, or staging) | false |
| `--no-color` |  | `false` | Disable colorful output | false |
| `--output` | `-o` | `table` | Output format (table, json, csv, xml) | false |


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

