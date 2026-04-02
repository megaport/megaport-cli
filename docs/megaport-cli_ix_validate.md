# validate

Validate an IX order without purchasing

## Description

Validates an IX configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `asn`: ASN (Autonomous System Number) for BGP peering
  - `mac-address`: MAC address for the IX interface
  - `name`: The name of the IX
  - `network-service-type`: The IX type/network service to connect to
  - `product-uid`: The UID of the port to attach the IX to
  - `rate-limit`: Rate limit in Mbps
  - `vlan`: VLAN ID for the IX connection

### Optional Fields
  - `promo-code`: Optional promotion code for discounts
  - `shutdown`: Whether the IX is initially shut down

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli ix validate --product-uid port-uid --name "My IX" --network-service-type "Los Angeles IX" --asn 65000 --mac-address "00:11:22:33:44:55" --rate-limit 1000 --vlan 100
  megaport-cli ix validate --json-file ./ix-config.json
```

## Usage

```sh
megaport-cli ix validate [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--asn` |  | `0` | ASN (Autonomous System Number) for BGP peering | true |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--mac-address` |  |  | MAC address for the IX interface | true |
| `--name` |  |  | The name of the IX | true |
| `--network-service-type` |  |  | The IX type/network service to connect to | true |
| `--product-uid` |  |  | The UID of the port to attach the IX to | true |
| `--promo-code` |  |  | Optional promotion code for discounts | false |
| `--rate-limit` |  | `0` | Rate limit in Mbps | true |
| `--shutdown` |  | `false` | Whether the IX is initially shut down | false |
| `--vlan` |  | `0` | VLAN ID for the IX connection | true |

## Subcommands
* [docs](megaport-cli_ix_validate_docs.md)

