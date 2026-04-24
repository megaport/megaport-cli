# list

List all partner ports

## Description

List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports. You can filter the partner ports based on specific criteria.

### Important Notes
  - The list can be filtered by multiple criteria at once
  - Filtering is case-insensitive and partial matches are supported

### Example Usage

```sh
  megaport-cli partners list
  megaport-cli partners list --product-name "AWS Partner Port"
  megaport-cli partners list --connect-type "Dedicated Cloud Connection"
  megaport-cli partners list --company-name "Amazon Web Services"
  megaport-cli partners list --location-id 1
  megaport-cli partners list --diversity-zone "blue"
```

## Usage

```sh
megaport-cli partners list [flags]
```


## Parent Command

* [megaport-cli partners](megaport-cli_partners.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--company-name` |  |  | Filter partner ports by company name | false |
| `--connect-type` |  |  | Filter partner ports by connect type | false |
| `--diversity-zone` |  |  | Filter partner ports by diversity zone | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--location-id` |  | `0` | Filter partner ports by location ID | false |
| `--product-name` |  |  | Filter partner ports by product name | false |

## Subcommands
* [docs](megaport-cli_partners_list_docs.md)

