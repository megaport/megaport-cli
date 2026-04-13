# megaport-cli

A CLI tool to interact with the Megaport API

## Description

Megaport CLI provides a command line interface to interact with the Megaport API.

The CLI allows you to manage Megaport resources such as ports, VXCs, MCRs, MVEs, service keys, and more.

### Optional Fields
  - `--env`: Environment to use (production, staging, development)
  - `--help`: Show help for any command
  - `--max-retries`: Maximum number of retries for transient API failures (default 3)
  - `--no-color`: Disable colored output
  - `--no-retry`: Disable automatic retry on transient API failures
  - `--output`: Output format (json, yaml, table, csv, xml)
  - `--quiet`: Suppress informational output, only show errors and data
  - `--verbose`: Show additional debug information

### Important Notes
  - Use the --help flag with any command to see specific usage information
  - Authentication is handled via the MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY environment variables
  - By default, the CLI connects to the Megaport production environment
  - Set the MEGAPORT_ENVIRONMENT environment variable to connect to a different environment

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
| `--fields` |  |  | Comma-separated list of fields to include in output (e.g., uid,name,status); use an unknown name to list available fields | false |
| `--log-http` |  | `false` | Log raw HTTP requests/responses to stderr for debugging (may include sensitive data such as auth tokens) | false |
| `--max-retries` |  | `3` | Maximum number of retries for transient API failures | false |
| `--no-color` |  | `false` | Disable colorful output | false |
| `--no-retry` |  | `false` | Disable automatic retry on transient API failures | false |
| `--output` | `-o` | `table` | Output format (table, json, csv, xml) | false |
| `--profile` |  |  | Use a specific config profile for this command | false |
| `--query` |  |  | JMESPath query to filter JSON output (requires --output json) | false |
| `--quiet` | `-q` | `false` | Suppress informational output, only show errors and data | false |
| `--timeout` |  | `0s` | Request timeout duration (e.g., 30s, 2m, 5m); 0 uses the internal default of 90s | false |
| `--verbose` | `-v` | `false` | Show additional debug information | false |

## Subcommands
* [apply](megaport-cli_apply.md)
* [auth](megaport-cli_auth.md)
* [billing-market](megaport-cli_billing-market.md)
* [completion](megaport-cli_completion.md)
* [config](megaport-cli_config.md)
* [generate-docs](megaport-cli_generate-docs.md)
* [ix](megaport-cli_ix.md)
* [locations](megaport-cli_locations.md)
* [managed-account](megaport-cli_managed-account.md)
* [mcr](megaport-cli_mcr.md)
* [mve](megaport-cli_mve.md)
* [partners](megaport-cli_partners.md)
* [ports](megaport-cli_ports.md)
* [product](megaport-cli_product.md)
* [servicekeys](megaport-cli_servicekeys.md)
* [status](megaport-cli_status.md)
* [topology](megaport-cli_topology.md)
* [users](megaport-cli_users.md)
* [version](megaport-cli_version.md)
* [vxc](megaport-cli_vxc.md)
* [whoami](megaport-cli_whoami.md)

