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

### Important Notes
  - Required flags can be skipped when using --interactive, --json, or --json-file


## Usage

```sh
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
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--lag-count` |  | `0` | The number of LAG members (between 1 and 8) | true |
| `--location-id` |  | `0` | The ID of the location where the port will be provisioned | true |
| `--marketplace-visibility` |  | `false` | Whether the port should be visible in the marketplace (true or false) | true |
| `--name` |  |  | The name of the port (1-64 characters) | true |
| `--port-speed` |  | `0` | The speed of each LAG member port (10000 or 100000 Mbps) | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | The term of the port (1, 12, or 24 months) | true |


