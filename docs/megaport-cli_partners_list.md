# list

List all partner ports

## Description

List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports. You can filter
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
  megaport-cli partners list --connect-type "Dedicated Cloud Connection"
  megaport-cli partners list --company-name "Amazon Web Services"
  megaport-cli partners list --location-id 1
  megaport-cli partners list --diversity-zone "Zone A"

```

Example output:
```
  Product Name        Connect Type              Company Name          Location ID  Diversity Zone
  ------------------  ------------------------  --------------------  -----------  --------------
  AWS Partner Port    Dedicated Cloud Connect   Amazon Web Services             1  Zone A

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



