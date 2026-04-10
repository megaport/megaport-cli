# get

Get details for a single MCR

## Description

Get details for a single MCR.

This command retrieves and displays detailed information for a single Megaport Cloud Router (MCR). You must provide the unique identifier (UID) of the MCR you wish to retrieve.

### Important Notes
  - The output includes the MCR's UID, name, location ID, port speed, and provisioning status

### Example Usage

```sh
  megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef --export
  megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef --watch
  megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef --watch --interval 10s
```

## Usage

```sh
megaport-cli mcr get [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)

## Aliases

* show
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--export` |  | `false` | Output recreatable JSON config for use with buy --json (excludes read-only fields) | false |
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

## Subcommands
* [docs](megaport-cli_mcr_get_docs.md)

