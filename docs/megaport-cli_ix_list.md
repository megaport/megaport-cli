# list

List all IXs with optional filters

## Description

List all IXs available in the Megaport API.

This command fetches and displays a list of IXs with details such as UID, name, network service type, ASN, rate limit, VLAN, and status. By default, only active IXs are shown.

### Optional Fields
  - `asn`: Filter IXs by ASN
  - `include-inactive`: Include IXs in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states
  - `location-id`: Filter IXs by location ID
  - `name`: Filter IXs by name (partial match)
  - `network-service-type`: Filter IXs by network service type
  - `rate-limit`: Filter IXs by rate limit in Mbps
  - `vlan`: Filter IXs by VLAN

### Example Usage

```sh
  megaport-cli ix list
  megaport-cli ix list --name "My IX"
  megaport-cli ix list --asn 65000
  megaport-cli ix list --network-service-type "Los Angeles IX"
  megaport-cli ix list --include-inactive
```

## Usage

```sh
megaport-cli ix list [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--asn` |  | `0` | Filter IXs by ASN | false |
| `--include-inactive` |  | `false` | Include inactive IXs in the list | false |
| `--location-id` |  | `0` | Filter IXs by location ID | false |
| `--name` |  |  | Filter IXs by name (partial match) | false |
| `--network-service-type` |  |  | Filter IXs by network service type | false |
| `--rate-limit` |  | `0` | Filter IXs by rate limit in Mbps | false |
| `--vlan` |  | `0` | Filter IXs by VLAN | false |

## Subcommands
* [docs](megaport-cli_ix_list_docs.md)

