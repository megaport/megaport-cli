# create

Create a new service key

## Description

Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.

Required fields:
product-uid: Product UID for the service key

Optional fields:
product-id: Product ID for the service key
single-use: Single-use service key
max-speed: Maximum speed for the service key
description: Description for the service key
start-date: Start date for the service key (YYYY-MM-DD)
end-date: End date for the service key (YYYY-MM-DD)

Example usage:

create --product-uid "product-uid" --description "My service key"
create --product-uid "product-uid" --single-use --max-speed 1000 --description "Single-use key"
create --product-uid "product-uid" --start-date "2023-01-01" --end-date "2023-12-31"



## Usage

```
megaport-cli servicekeys create [flags]
```



## Parent Command

* [megaport-cli servicekeys](megaport-cli_servicekeys.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--description` |  |  | Description for the service key | false |
| `--end-date` |  |  | End date for the service key (YYYY-MM-DD) | false |
| `--max-speed` |  | `0` | Maximum speed for the service key | false |
| `--product-id` |  | `0` | Product ID for the service key | false |
| `--product-uid` |  |  | Product UID for the service key | false |
| `--single-use` |  | `false` | Single-use service key | false |
| `--start-date` |  |  | Start date for the service key (YYYY-MM-DD) | false |



