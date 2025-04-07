# buy-lag

Buy a LAG port through the Megaport API

## Description

Buy a LAG port through the Megaport API.

This command allows you to purchase a LAG port by providing the necessary details.

### Required Fields
  - `lag-count`: The number of LAG members (between 1 and 8)
  - `location-id`: The ID of the location where the port will be provisioned
  - `marketplace-visibility`: Whether the port should be visible in the marketplace (true or false)
  - `name`: The name of the port (1-64 characters)
  - `port-speed`: The speed of each LAG member port (10000 or 100000 Mbps)
  - `term`: The term of the port (1, 12, or 24 months)

### Optional Fields
  - `cost-centre`: The cost centre for the LAG port
  - `diversity-zone`: The diversity zone for the LAG port
  - `promo-code`: A promotional code for the LAG port

### Example Usage

```
  buy-lag --interactive
  buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
  buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
  buy-lag --json-file ./lag-port-config.json
```
### JSON Format Example
```json
{
  "name": "My LAG Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "lagCount": 2,
  "marketPlaceVisibility": true,
  "diversityZone": "A",
  "costCentre": "IT-2023"
}

```


## Usage

```
megaport-cli ports buy-lag [flags]
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
| `--lag-count` |  | `0` | Number of LAGs (1-8) | false |
| `--location-id` |  | `0` | Location ID where the port will be provisioned | false |
| `--marketplace-visibility` |  | `false` | Whether the port is visible in marketplace | false |
| `--name` |  |  | Port name | false |
| `--port-speed` |  | `0` | Port speed in Mbps (10000 or 100000) | false |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | false |



