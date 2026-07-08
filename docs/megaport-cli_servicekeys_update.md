# update

Update an existing service key

## Description

Update an existing service key for the Megaport API.

This command allows you to modify the details of an existing service key. You need to specify the key identifier as an argument, and provide any updated values as flags.

### Important Notes
  - Only specified fields will be updated; unspecified fields will remain unchanged

### Example Usage

```sh
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --active
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --active=false
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --product-uid "new-product-uid"
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --interactive
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --json '{"active":false}'
```
### JSON Format Example
```json
{
  "productUid": "new-product-uid",
  "singleUse": false,
  "active": false
}

```

## Usage

```sh
megaport-cli servicekeys update [flags]
```


## Parent Command

* [megaport-cli servicekeys](megaport-cli_servicekeys.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--active` |  | `false` | Activate or deactivate the service key | false |
| `--generate-skeleton` |  | `false` | Print a JSON skeleton template for --json or --json-file input and exit | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--product-id` |  | `0` | Product ID for the service key | false |
| `--product-uid` |  |  | Product UID for the service key | false |
| `--single-use` |  | `false` | Single-use service key | false |

