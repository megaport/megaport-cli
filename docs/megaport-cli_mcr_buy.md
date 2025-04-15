# buy

Buy an MCR through the Megaport API

## Description

Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.

### Required Fields
  - `location-id`: The ID of the location where the MCR will be provisioned
  - `marketplace-visibility`: Whether the MCR should be visible in the marketplace (true or false)
  - `name`: The name of the MCR (1-64 characters)
  - `port-speed`: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps)
  - `term`: The term of the MCR (1, 12, 24, or 36 months)

### Optional Fields
  - `cost-centre`: The cost centre for the MCR
  - `diversity-zone`: The diversity zone for the MCR
  - `mcr-asn`: The ASN for the MCR (if not provided, a private ASN will be assigned)
  - `promo-code`: A promotional code for the MCR
  - `resource-tags`: JSON string of key-value pairs for resource tagging
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - The location_id must correspond to a valid location in the Megaport API
  - The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps)
  - If mcr_asn is not provided, a private ASN will be automatically assigned
  - Resource tags allow you to categorize resources for organization and billing purposes
  - Required flags (name, term, port-speed, location-id, marketplace-visibility) can be skipped when using --interactive, --json, or --json-file

### Example Usage

```sh
  megaport-cli mcr buy --interactive
  megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --marketplace-visibility true --mcr-asn 65000
  megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --marketplace-visibility true --resource-tags '{"env":"prod","owner":"network-team"}'
  megaport-cli mcr buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123,"mcrAsn":65000,"marketplaceVisibility":true}'
  megaport-cli mcr buy --json-file ./mcr-config.json
```
### JSON Format Example
```json
{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "marketplaceVisibility": true,
  "diversityZone": "blue",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024",
  "resourceTags": {
    "environment": "production",
    "department": "networking",
    "project": "cloud-migration",
    "owner": "john.doe@example.com"
  }
}

```

## Usage

```sh
megaport-cli mcr buy [flags]
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
| `--port-speed` |  | `0` | The speed of the MCR (1000, 2500, 5000, or 10000 Mbps) | true |
| `--promo-code` |  |  | A promotional code for discounts | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--term` |  | `0` | The term of the MCR (1, 12, 24, or 36 months) | true |

## Subcommands
* [docs](megaport-cli_mcr_buy_docs.md)

