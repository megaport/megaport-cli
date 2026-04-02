# validate

Validate an MCR order without purchasing

## Description

Validates an MCR configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `location-id`: The ID of the location where the MCR will be provisioned
  - `marketplace-visibility`: Whether the MCR should be visible in the marketplace (true or false)
  - `name`: The name of the MCR (1-64 characters)
  - `port-speed`: The speed of the MCR (1000, 2500, 5000, 10000, 25000, 50000, or 100000 Mbps)
  - `term`: The term of the MCR (1, 12, 24, or 36 months)

### Optional Fields
  - `cost-centre`: The cost centre for billing purposes
  - `diversity-zone`: The diversity zone for the MCR
  - `mcr-asn`: The ASN for the MCR (64512-65534 for private ASN, or a public ASN)
  - `promo-code`: A promotional code for discounts
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli mcr validate --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --marketplace-visibility true
  megaport-cli mcr validate --json-file ./mcr-config.json
```

## Usage

```sh
megaport-cli mcr validate [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | The cost centre for billing purposes | false |
| `--diversity-zone` |  |  | The diversity zone for the MCR | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The ID of the location where the MCR will be provisioned | true |
| `--marketplace-visibility` |  |  | Whether the MCR should be visible in the marketplace (true or false) | true |
| `--mcr-asn` |  | `0` | The ASN for the MCR (64512-65534 for private ASN, or a public ASN) | false |
| `--name` |  |  | The name of the MCR (1-64 characters) | true |
| `--port-speed` |  | `0` | The speed of the MCR (1000, 2500, 5000, 10000, 25000, 50000, or 100000 Mbps) | true |
| `--promo-code` |  |  | A promotional code for discounts | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--term` |  | `0` | The term of the MCR (1, 12, 24, or 36 months) | true |

## Subcommands
* [docs](megaport-cli_mcr_validate_docs.md)

