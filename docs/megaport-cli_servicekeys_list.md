# list

List all service keys

## Description

List all service keys for the Megaport API.

This command retrieves and displays all service keys along with their details. Use this command to review the keys available in your account.

### Example Usage

```sh
  megaport-cli servicekeys list
  megaport-cli servicekeys list --product-uid "product-uid"
```

## Usage

```sh
megaport-cli servicekeys list [flags]
```


## Parent Command

* [megaport-cli servicekeys](megaport-cli_servicekeys.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--product-uid` |  |  | Filter service keys by product UID | false |

## Subcommands
* [docs](megaport-cli_servicekeys_list_docs.md)

