# buy

Buy an MCR through the Megaport API

## Description

Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --term, --port-speed, --location-id

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
  - name: The name of the MCR.
  - term: The term of the MCR (1, 12, 24, or 36 months).
  - port_speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
  - location_id: The ID of the location where the MCR will be provisioned.

Optional fields:
  - mcr_asn: The ASN for the MCR.
  - diversity_zone: The diversity zone for the MCR.
  - cost_centre: The cost center for the MCR.
  - promo_code: A promotional code for the MCR.
  - resource_tags: Key-value tags to associate with the MCR (JSON format).

Example usage:

```
  # Interactive mode
  megaport-cli mcr buy --interactive
```

```
  # Flag mode
  megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123
```

```
  # JSON mode
  megaport-cli mcr buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123}'
  megaport-cli mcr buy --json-file ./mcr-config.json
```



## Usage

```
megaport-cli mcr buy [flags]
```



## Parent Command

* [megaport-cli mcr](mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --cost-centre |  |  | Cost centre for billing | false |
| --diversity-zone |  |  | Diversity zone for the MCR | false |
| --interactive | -i | false | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing MCR configuration | false |
| --json-file |  |  | Path to JSON file containing MCR configuration | false |
| --location-id |  | 0 | Location ID where the MCR will be provisioned | false |
| --mcr-asn |  | 0 | ASN for the MCR (optional) | false |
| --name |  |  | MCR name | false |
| --port-speed |  | 0 | Port speed in Mbps (1000, 2500, 5000, or 10000) | false |
| --promo-code |  |  | Promotional code for discounts | false |
| --resource-tags |  |  | JSON string of key-value resource tags | false |
| --term |  | 0 | Contract term in months (1, 12, 24, or 36) | false |



