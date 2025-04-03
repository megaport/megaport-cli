# buy

Buy an MCR through the Megaport API

## Description

Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field, guiding you through the configuration process.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --term, --port-speed, --location-id
   Optional fields can also be specified using flags.

3. JSON Mode:
   Provide a JSON string or file with all required and optional fields:
   --json <json-string> or --json-file <path>

Required fields:
- `name`: The name of the MCR (1-64 characters).
- `term`: The contract term for the MCR (1, 12, 24, or 36 months).
- `port_speed`: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
- `location_id`: The ID of the location where the MCR will be provisioned.

Optional fields:
- `mcr_asn`: The ASN for the MCR (64512-65534 for private ASN, or a public ASN). If not provided, a private ASN will be automatically assigned.
- `diversity_zone`: The diversity zone for the MCR (if applicable).
- `cost_centre`: The cost center for billing purposes.
- `promo_code`: A promotional code for discounts (if applicable).

Example usage:

### Interactive mode
```
  megaport-cli mcr buy --interactive

```

### Flag mode
```
  megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --mcr-asn 65000 --resource-tags '{"environment":"production"}'

```

### JSON mode
```
  megaport-cli mcr buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123,"mcrAsn":65000,"resourceTags":{"environment":"production"}}'
  megaport-cli mcr buy --json-file ./mcr-config.json

```

JSON format example (mcr-config.json):
```
{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "diversityZone": "zone-a",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024",
}

```

Notes:
- The location_id must correspond to a valid location in the Megaport API.
- The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps).
- If mcr_asn is not provided, a private ASN will be automatically assigned.



## Usage

```
megaport-cli mcr buy [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | Diversity zone for the MCR | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing MCR configuration | false |
| `--json-file` |  |  | Path to JSON file containing MCR configuration | false |
| `--location-id` |  | `0` | Location ID where the MCR will be provisioned | false |
| `--mcr-asn` |  | `0` | ASN for the MCR (optional) | false |
| `--name` |  |  | MCR name | false |
| `--port-speed` |  | `0` | Port speed in Mbps (1000, 2500, 5000, or 10000) | false |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | false |



