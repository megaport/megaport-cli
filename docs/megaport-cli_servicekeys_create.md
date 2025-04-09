# create

Create a new service key

## Description

Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.

### Example Usage

```sh
  megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"
  megaport-cli servicekeys create --product-uid "product-uid" --single-use --max-speed 1000 --description "Single-use key"
  megaport-cli servicekeys create --product-uid "product-uid" --start-date "2023-01-01" --end-date "2023-12-31"
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
| `--description` |  |  | Description for the service key | false |
| `--end-date` |  |  | End date (YYYY-MM-DD) | false |
| `--max-speed` |  | `0` | Maximum speed for the service key | false |
| `--product-id` |  | `0` | Product ID for the service key | false |
| `--product-uid` |  |  | Product UID for the service key | false |
| `--single-use` |  | `false` | Single-use service key | false |
| `--start-date` |  |  | Start date (YYYY-MM-DD) | false |

