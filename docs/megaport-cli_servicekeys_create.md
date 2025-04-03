# create

Create a new service key

## Description

Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.
You may need to provide additional flags or parameters based on your API requirements.

Example:
```
megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"

Example output:
Key: a1b2c3d4-e5f6-7890-1234-567890abcdef  Product UID: product-uid  Description: My service key

```


## Usage

```
megaport-cli servicekeys create [flags]
```

## Examples

```
Example:
megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"

Example output:
  Key: a1b2c3d4-e5f6-7890-1234-567890abcdef  Product UID: product-uid  Description: My service key
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



