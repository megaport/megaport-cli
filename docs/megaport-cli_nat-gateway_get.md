# get

Get details for a single NAT Gateway

## Description

Get details for a single NAT Gateway.

Retrieves and displays detailed information for a single NAT Gateway by its product UID.

### Example Usage

```sh
  megaport-cli nat-gateway get a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli nat-gateway get a1b2c3d4-e5f6-7890-1234-567890abcdef --export
  megaport-cli nat-gateway get a1b2c3d4-e5f6-7890-1234-567890abcdef --watch
```

## Usage

```sh
megaport-cli nat-gateway get [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)

## Aliases

* show
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--export` |  | `false` | Output recreatable JSON config for use with create --json-file | false |
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

