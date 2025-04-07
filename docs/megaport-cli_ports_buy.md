# buy

Buy a port through the Megaport API

## Description

Buy a port through the Megaport API.

This command allows you to purchase a port by providing the necessary details.

Required fields:
  location-id: The ID of the location where the port will be provisioned
  marketplace-visibility: Whether the port should be visible in the marketplace (true or false)
  name: The name of the port (1-64 characters)
  port-speed: The speed of the port (1000, 10000, or 100000 Mbps)
  term: The term of the port (1, 12, 24, or 36 months)

Optional fields:
  cost-centre: The cost center for the port
  diversity-zone: The diversity zone for the port
  promo-code: A promotional code for the port

Example usage:

```
  buy --interactive
  buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
  buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
  buy --json-file ./port-config.json
```
JSON format example:
```
{
  "name": "My Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "marketPlaceVisibility": true,
  "diversityZone": "A",
  "costCentre": "IT-2023"
}
```


## Usage

```
megaport-cli ports buy [flags]
```



## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | Diversity zone for the port | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing port configuration | false |
| `--json-file` |  |  | Path to JSON file containing port configuration | false |
| `--location-id` |  | `0` | Location ID where the port will be provisioned | false |
| `--marketplace-visibility` |  | `false` | Whether the port is visible in marketplace | false |
| `--name` |  |  | Port name | false |
| `--port-speed` |  | `0` | Port speed in Mbps (1000, 10000, or 100000) | false |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | false |



