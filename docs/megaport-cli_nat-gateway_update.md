# update

Update an existing NAT Gateway

## Description

Update an existing NAT Gateway.

This command allows you to update the details of an existing NAT Gateway.

### Optional Fields
  - `auto-renew`: Whether to automatically renew the contract term
  - `diversity-zone`: The new diversity zone
  - `location-id`: The new location ID
  - `name`: The new name of the NAT Gateway
  - `promo-code`: A promotional code
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags
  - `service-level-reference`: A service level reference
  - `session-count`: The new session count
  - `speed`: The new speed of the NAT Gateway in Mbps
  - `term`: The new contract term in months

### Example Usage

```sh
  megaport-cli nat-gateway update [uid] --interactive
  megaport-cli nat-gateway update [uid] --name "Updated GW" --speed 2000
  megaport-cli nat-gateway update [uid] --json '{"name":"Updated GW","speed":2000,"locationId":123,"term":12}'
  megaport-cli nat-gateway update [uid] --json-file ./update-config.json
```
### JSON Format Example
```json
{
  "name": "Updated NAT Gateway",
  "term": 12,
  "speed": 2000,
  "locationId": 123,
  "sessionCount": 200
}

```

## Usage

```sh
megaport-cli nat-gateway update [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--auto-renew` |  | `false` | Whether to automatically renew the contract term | false |
| `--diversity-zone` |  |  | The new diversity zone | false |
| `--generate-skeleton` |  | `false` | Print a JSON skeleton template for --json or --json-file input and exit | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The new location ID | false |
| `--name` |  |  | The new name of the NAT Gateway | false |
| `--promo-code` |  |  | A promotional code | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--service-level-reference` |  |  | A service level reference | false |
| `--session-count` |  | `0` | The new session count | false |
| `--speed` |  | `0` | The new speed of the NAT Gateway in Mbps | false |
| `--term` |  | `0` | The new contract term in months (1, 12, 24, 36, 48, or 60) | false |

