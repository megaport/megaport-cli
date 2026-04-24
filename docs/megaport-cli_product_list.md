# list

List all products with optional filters

## Description

List all products available in the Megaport API.

This command fetches and displays a list of products with details such as UID, name, type, location, speed, and status. By default, only active products are shown.

### Optional Fields
  - `include-inactive`: Include products in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states

### Example Usage

```sh
  megaport-cli product list
  megaport-cli product list --include-inactive
  megaport-cli product list --limit 10
```

## Usage

```sh
megaport-cli product list [flags]
```


## Parent Command

* [megaport-cli product](megaport-cli_product.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include products in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |

## Subcommands
* [docs](megaport-cli_product_list_docs.md)

