# validate

Validate a VXC order without purchasing

## Description

Validates a VXC configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `a-end-uid`: UID of the A-End product
  - `name`: Name of the VXC
  - `rate-limit`: Bandwidth in Mbps
  - `term`: Contract term in months (1, 12, 24, or 36)

### Optional Fields
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli vxc validate --name "My VXC" --rate-limit 1000 --term 12 --a-end-uid port-123 --b-end-uid port-456 --a-end-vlan 100 --b-end-vlan 200
  megaport-cli vxc validate --json-file ./vxc-config.json
```

## Usage

```sh
megaport-cli vxc validate [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--a-end-inner-vlan` |  | `0` | Inner VLAN for A-End (-1 or higher) | false |
| `--a-end-partner-config` |  |  | JSON string with A-End partner configuration | false |
| `--a-end-uid` |  |  | UID of the A-End product | true |
| `--a-end-vlan` |  | `0` | VLAN for A-End (0-4093, except 1) | false |
| `--a-end-vnic-index` |  | `0` | vNIC index for A-End MVE | false |
| `--b-end-inner-vlan` |  | `0` | Inner VLAN for B-End (-1 or higher) | false |
| `--b-end-partner-config` |  |  | JSON string with B-End partner configuration | false |
| `--b-end-uid` |  |  | UID of the B-End product | false |
| `--b-end-vlan` |  | `0` | VLAN for B-End (0-4093, except 1) | false |
| `--b-end-vnic-index` |  | `0` | vNIC index for B-End MVE | false |
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--name` |  |  | Name of the VXC | true |
| `--promo-code` |  |  | Promotional code | false |
| `--rate-limit` |  | `0` | Bandwidth in Mbps | true |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--service-key` |  |  | Service key | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | true |

## Subcommands
* [docs](megaport-cli_vxc_validate_docs.md)

