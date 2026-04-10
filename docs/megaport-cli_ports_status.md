# status

Check the provisioning status of a port

## Description

Check the provisioning status of a port through the Megaport API.

This command retrieves only the essential status information for a port without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the port's status without retrieving all details.

### Example Usage

```sh
  megaport-cli ports status port-abc123
  megaport-cli ports status port-abc123 --watch
  megaport-cli ports status port-abc123 --watch --interval 10s
```

## Usage

```sh
megaport-cli ports status [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)

## Aliases

* st
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

## Subcommands
* [docs](megaport-cli_ports_status_docs.md)

