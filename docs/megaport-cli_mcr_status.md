# status

Check the provisioning status of an MCR

## Description

Check the provisioning status of an MCR through the Megaport API.

This command retrieves only the essential status information for a Megaport Cloud Router (MCR) without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the MCR's status without retrieving all details.

### Example Usage

```sh
  megaport-cli mcr status mcr-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  megaport-cli mcr status mcr-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --watch
  megaport-cli mcr status mcr-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --watch --interval 10s
```

## Usage

```sh
megaport-cli mcr status [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)

## Aliases

* st
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

## Subcommands
* [docs](megaport-cli_mcr_status_docs.md)

