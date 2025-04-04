# list

List all partner ports

## Description

List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports. You can filter the partner ports based on specific criteria.

Optional fields:
location-id: Filter partner ports by location ID
diversity-zone: Filter partner ports by diversity zone
product-name: Filter partner ports by product name
connect-type: Filter partner ports by connect type
company-name: Filter partner ports by company name

Important notes:
- The list can be filtered by multiple criteria at once
- Filtering is case-insensitive and partial matches are supported

Example usage:

list
list --product-name "AWS Partner Port"
list --connect-type "Dedicated Cloud Connection"
list --company-name "Amazon Web Services"
list --location-id 1
list --diversity-zone "Zone A"



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



