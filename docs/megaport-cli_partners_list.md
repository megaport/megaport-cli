# list

List all partner ports

## Description

List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports with details such as
product name, connect type, company name, location ID, and diversity zone. You can also filter
the partner ports based on specific criteria.

Available filters:
- `product-name`: Filter partner ports by product name.
- `connect-type`: Filter partner ports by connect type.
- `company-name`: Filter partner ports by company name.
- `location-id`: Filter partner ports by location ID.
- `diversity-zone`: Filter partner ports by diversity zone.

Example usage:

```
  megaport-cli partners list
  megaport-cli partners list --product-name "AWS Partner Port"
  megaport-cli partners list --connect-type "AWS"
  megaport-cli partners list --company-name "AWS"
  megaport-cli partners list --location-id 67
  megaport-cli partners list --diversity-zone "blue"

```



## Usage

```
megaport-cli partners list [flags]
```



## Parent Command

* [megaport-cli partners](megaport-cli_partners.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--company-name` |  |  | Filter by Company Name | false |
| `--connect-type` |  |  | Filter by Connect Type | false |
| `--diversity-zone` |  |  | Filter by Diversity Zone | false |
| `--location-id` |  | `0` | Filter by Location ID | false |
| `--product-name` |  |  | Filter by Product Name | false |



