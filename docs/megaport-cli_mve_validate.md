# validate

Validate an MVE order without purchasing

## Description

Validates an MVE configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `location-id`: The ID of the location where the MVE will be provisioned
  - `name`: The name of the MVE
  - `term`: The term of the MVE (1, 12, 24, or 36 months)
  - `vendor-config`: JSON string with vendor-specific configuration
  - `vnics`: JSON array of network interfaces

### Optional Fields
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli mve validate --name "My MVE" --term 12 --location-id 123 --vendor-config '{"vendor":"cisco","imageId":123,"productSize":"MEDIUM"}' --vnics '[{"description":"Data Plane","vlan":100}]'
  megaport-cli mve validate --json-file ./mve-config.json
```

## Usage

```sh
megaport-cli mve validate [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | The diversity zone for the MVE | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The ID of the location where the MVE will be provisioned | true |
| `--name` |  |  | The name of the MVE | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--term` |  | `0` | The term of the MVE (1, 12, 24, or 36 months) | true |
| `--vendor-config` |  |  | JSON string with vendor-specific configuration | true |
| `--vnics` |  |  | JSON array of network interfaces | true |

## Subcommands
* [docs](megaport-cli_mve_validate_docs.md)

