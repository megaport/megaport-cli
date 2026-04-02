# update

Update an existing IX

## Description

Update an existing Internet Exchange (IX).

This command allows you to update the details of an existing IX.

### Optional Fields
  - `a-end-product-uid`: Move the IX by changing the A-End of the IX
  - `asn`: ASN (Autonomous System Number) for BGP peering
  - `cost-centre`: Cost centre for invoicing purposes
  - `mac-address`: MAC address for the IX interface
  - `name`: The new name of the IX
  - `password`: BGP password
  - `public-graph`: Whether the IX usage statistics are publicly viewable
  - `rate-limit`: Rate limit in Mbps
  - `reverse-dns`: DNS lookup of a domain name from an IP address
  - `shutdown`: Shut down or re-enable the IX
  - `vlan`: VLAN ID for the IX connection

### Important Notes
  - The IX UID cannot be changed
  - Only specified fields will be updated; unspecified fields will remain unchanged

### Example Usage

```sh
  megaport-cli ix update [ixUID] --interactive
  megaport-cli ix update [ixUID] --name "Updated IX" --rate-limit 2000
  megaport-cli ix update [ixUID] --json '{"name":"Updated IX","rateLimit":2000}'
  megaport-cli ix update [ixUID] --json-file ./update-ix-config.json
```
### JSON Format Example
```json
{
  "name": "Updated IX",
  "rateLimit": 2000,
  "costCentre": "Finance",
  "vlan": 200,
  "macAddress": "00:11:22:33:44:66",
  "asn": 65001
}

```

## Usage

```sh
megaport-cli ix update [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--a-end-product-uid` |  |  | Move the IX by changing the A-End of the IX | false |
| `--asn` |  | `0` | ASN (Autonomous System Number) for BGP peering | false |
| `--cost-centre` |  |  | Cost centre for invoicing purposes | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--mac-address` |  |  | MAC address for the IX interface | false |
| `--name` |  |  | The new name of the IX | false |
| `--password` |  |  | BGP password | false |
| `--public-graph` |  | `false` | Whether the IX usage statistics are publicly viewable | false |
| `--rate-limit` |  | `0` | Rate limit in Mbps | false |
| `--reverse-dns` |  |  | DNS lookup of a domain name from an IP address | false |
| `--shutdown` |  | `false` | Shut down or re-enable the IX | false |
| `--vlan` |  | `0` | VLAN ID for the IX connection | false |

## Subcommands
* [docs](megaport-cli_ix_update_docs.md)

