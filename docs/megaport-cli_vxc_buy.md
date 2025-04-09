# buy

Purchase a new VXC

## Description

Purchase a new Megaport Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to create a VXC by providing the necessary details.

### Required Fields
  - `a-end-uid`: UID of the A-End product
  - `a-end-vlan`: VLAN for A-End (0-4093, except 1)
  - `b-end-uid`: UID of the B-End product (if not using partner configuration)
  - `b-end-vlan`: VLAN for B-End (0-4093, except 1)
  - `name`: Name of the VXC
  - `rate-limit`: Bandwidth in Mbps
  - `term`: Contract term in months (1, 12, 24, or 36)

### Example Usage

```sh
  megaport-cli vxc buy --interactive
  megaport-cli vxc buy --name "My VXC" --rate-limit 1000 --term 12 --a-end-uid port-123 --b-end-uid port-456 --a-end-vlan 100 --b-end-vlan 200
  megaport-cli vxc buy --json '{"vxcName":"My VXC","rateLimit":1000,"term":12,"portUid":"port-123","aEndConfiguration":{"vlan":100},"bEndConfiguration":{"productUID":"port-456","vlan":200}}'
  megaport-cli vxc buy --json-file ./vxc-config.json
```
### JSON Format Example
```json
{
  "vxcName": "My VXC",
  "rateLimit": 1000, 
  "term": 12,
  "portUid": "port-123",
  "aEndConfiguration": {
    "vlan": 100
  },
  "bEndConfiguration": {
    "productUID": "port-456",
    "vlan": 200
  },
  "costCentre": "IT Department"
}

```

## Usage

```sh
megaport-cli vxc buy [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--a-end-inner-vlan` |  | `0` | Inner VLAN for A-End (-1 or higher) | false |
| `--a-end-partner-config` |  |  | JSON string with A-End partner configuration | false |
| `--a-end-uid` |  |  | UID of the A-End product | true |
| `--a-end-vlan` |  | `0` | VLAN for A-End (0-4093, except 1) | true |
| `--a-end-vnic-index` |  | `0` | vNIC index for A-End MVE | false |
| `--b-end-inner-vlan` |  | `0` | Inner VLAN for B-End (-1 or higher) | false |
| `--b-end-partner-config` |  |  | JSON string with B-End partner configuration | false |
| `--b-end-uid` |  |  | UID of the B-End product (if not using partner configuration) | true |
| `--b-end-vlan` |  | `0` | VLAN for B-End (0-4093, except 1) | true |
| `--b-end-vnic-index` |  | `0` | vNIC index for B-End MVE | false |
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--name` |  |  | Name of the VXC | true |
| `--promo-code` |  |  | Promotional code | false |
| `--rate-limit` |  | `0` | Bandwidth in Mbps | true |
| `--service-key` |  |  | Service key | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | true |

## Subcommands
* [docs](megaport-cli_vxc_buy_docs.md)

