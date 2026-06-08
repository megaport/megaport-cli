# telemetry

Get telemetry data for a NAT Gateway

## Description

Get telemetry data for a NAT Gateway.

Retrieves metric samples (bits, packets, speed, etc.) for a NAT Gateway over a specified time window.

### Required Fields
  - `types`: Comma-separated telemetry types (e.g. BITS,PACKETS,SPEED)

### Optional Fields
  - `days`: Number of days of telemetry to retrieve (1-180)
  - `from`: Start time for telemetry (RFC3339); requires --to
  - `to`: End time for telemetry (RFC3339); requires --from

### Important Notes
  - Use --days for a rolling window, or --from/--to for an absolute range (they are mutually exclusive)

### Example Usage

```sh
  megaport-cli nat-gateway telemetry [uid] --types BITS --days 7
  megaport-cli nat-gateway telemetry [uid] --types BITS,PACKETS --from 2024-01-01T00:00:00Z --to 2024-01-07T00:00:00Z
  megaport-cli nat-gateway telemetry [uid] --types SPEED --days 30 --output json
```

## Usage

```sh
megaport-cli nat-gateway telemetry [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--days` |  | `0` | Number of days of telemetry to retrieve (1-180); mutually exclusive with --from/--to | false |
| `--from` |  |  | Start time for telemetry in RFC3339 format (e.g. 2024-01-01T00:00:00Z); use with --to | false |
| `--to` |  |  | End time for telemetry in RFC3339 format; use with --from | false |
| `--types` |  |  | Comma-separated telemetry types (e.g. BITS,PACKETS,SPEED) | true |

## Subcommands
* [docs](megaport-cli_nat-gateway_telemetry_docs.md)

