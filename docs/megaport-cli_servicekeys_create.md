# create

Create a new service key

## Description

Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.

### Important Notes
  - Provide either productUid or productId, not both

### Example Usage

```sh
  megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"
  megaport-cli servicekeys create --product-uid "product-uid" --single-use --max-speed 1000 --description "Single-use key"
  megaport-cli servicekeys create --product-uid "product-uid" --start-date "2023-01-01" --end-date "2023-12-31"
  megaport-cli servicekeys create --interactive
  megaport-cli servicekeys create --json '{"productUid":"product-uid","description":"My service key"}'
  megaport-cli servicekeys create --json-file ./servicekey-config.json
```
### JSON Format Example
```json
{
  "productUid": "product-uid",
  "singleUse": false,
  "maxSpeed": 1000,
  "description": "My service key",
  "active": true,
  "preApproved": false,
  "vlan": 100,
  "startDate": "2023-01-01",
  "endDate": "2023-12-31"
}

```

## Usage

```sh
megaport-cli servicekeys create [flags]
```


## Parent Command

* [megaport-cli servicekeys](megaport-cli_servicekeys.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--active` |  | `false` | Make the service key available immediately | false |
| `--description` |  |  | Description for the service key | false |
| `--end-date` |  |  | End date (YYYY-MM-DD) | false |
| `--generate-skeleton` |  | `false` | Print a JSON skeleton template for --json or --json-file input and exit | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--max-speed` |  | `0` | Maximum speed for the service key | false |
| `--pre-approved` |  | `false` | Pre-approve the service key for use | false |
| `--product-id` |  | `0` | Product ID for the service key | false |
| `--product-uid` |  |  | Product UID for the service key | false |
| `--single-use` |  | `false` | Single-use service key | false |
| `--start-date` |  |  | Start date (YYYY-MM-DD) | false |
| `--vlan` |  | `0` | VLAN ID for the service key (required for single-use keys) | false |

