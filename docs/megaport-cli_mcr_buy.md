# buy

Buy an MCR through the Megaport API

## Description

Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.

Required fields:
  location-id: The ID of the location where the MCR will be provisioned
  name: The name of the MCR (1-64 characters)
  port-speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps)
  term: The contract term for the MCR (1, 12, 24, or 36 months)

Optional fields:
  cost-centre: The cost center for billing purposes
  diversity-zone: The diversity zone for the MCR
  mcr-asn: The ASN for the MCR (64512-65534 for private ASN, or a public ASN)
  promo-code: A promotional code for discounts

Important notes:
  - The location_id must correspond to a valid location in the Megaport API
  - The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps)
  - If mcr_asn is not provided, a private ASN will be automatically assigned

Example usage:

```
  buy --interactive
  buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --mcr-asn 65000
  buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123,"mcrAsn":65000}'
  buy --json-file ./mcr-config.json
```
JSON format example:
```
{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "diversityZone": "zone-a",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024"
}
```


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



