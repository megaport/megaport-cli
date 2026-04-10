# rtt

Query round-trip times between locations

## Description

Query median round-trip times (RTT) between Megaport locations.

This command retrieves latency data between a source location and all other Megaport locations for a given month. Use this for network planning — choosing MCR locations and designing cross-connects based on latency requirements.

By default, returns data for the current month. Use --year and --month to query historical data.

### Required Fields
  - `src-location`: Source location ID

### Example Usage

```sh
  megaport-cli locations rtt --src-location 67
  megaport-cli locations rtt --src-location 67 --dst-location 3
  megaport-cli locations rtt --src-location 67 --year 2026 --month 3
  megaport-cli locations rtt --src-location 67 --output json
```

## Usage

```sh
megaport-cli locations rtt [flags]
```


## Parent Command

* [megaport-cli locations](megaport-cli_locations.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--dst-location` |  | `0` | Filter results to a specific destination location ID | false |
| `--month` |  | `0` | Month for RTT data (1-12, default: current month) | false |
| `--src-location` |  | `0` | Source location ID | true |
| `--year` |  | `0` | Year for RTT data (default: current year) | false |

## Subcommands
* [docs](megaport-cli_locations_rtt_docs.md)

