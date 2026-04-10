# get

Get details for a single MVE

## Description

Get details for a single MVE from the Megaport API.

This command retrieves and displays detailed information for a single Megaport Virtual Edge (MVE). You must provide the unique identifier (UID) of the MVE you wish to retrieve.

### Important Notes
  - The output includes the MVE's UID, name, vendor, version, status, and connectivity details

### Example Usage

```sh
  megaport-cli mve get a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli mve get a1b2c3d4-e5f6-7890-1234-567890abcdef --export
  megaport-cli mve get a1b2c3d4-e5f6-7890-1234-567890abcdef --watch
  megaport-cli mve get a1b2c3d4-e5f6-7890-1234-567890abcdef --watch --interval 10s
```

## Usage

```sh
megaport-cli mve get [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)

## Aliases

* show
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--export` |  | `false` | Output recreatable JSON config for use with buy --json (excludes read-only fields; vendorConfig not available from API) | false |
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

## Subcommands
* [docs](megaport-cli_mve_get_docs.md)

