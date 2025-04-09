# buy

Buy a port through the Megaport API

## Description

Buy a port through the Megaport API.

This command allows you to purchase a port by providing the necessary details.

### Required Fields
  - `location-id`: The ID of the location where the port will be provisioned
  - `marketplace-visibility`: Whether the port should be visible in the marketplace (true or false)
  - `name`: The name of the port (1-64 characters)
  - `port-speed`: The speed of the port (1000, 10000, or 100000 Mbps)
  - `term`: The term of the port (1, 12, 24, or 36 months)

### Optional Fields
  - `cost-centre`: The cost centre for the port
  - `diversity-zone`: The diversity zone for the port
  - `promo-code`: A promotional code for the port

### Important Notes
  - Required flags (name, term, port-speed, location-id, marketplace-visibility) can be skipped when using --interactive, --json, or --json-file

### Example Usage

```sh
  megaport-cli ports buy --interactive
  megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
  megaport-cli ports buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
  megaport-cli ports buy --json-file ./port-config.json
```
### JSON Format Example
```json
{
  "name": "My Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "marketPlaceVisibility": true,
  "diversityZone": "blue",
  "costCentre": "IT-2023"
}

```

## Usage

```sh
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
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The ID of the location where the port will be provisioned | true |
| `--marketplace-visibility` |  | `false` | Whether the port should be visible in the marketplace (true or false) | true |
| `--name` |  |  | The name of the port (1-64 characters) | true |
| `--port-speed` |  | `0` | The speed of the port (1000, 10000, or 100000 Mbps) | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | The term of the port (1, 12, 24, or 36 months) | true |

## Subcommands
* [docs](megaport-cli_ports_buy_docs.md)

