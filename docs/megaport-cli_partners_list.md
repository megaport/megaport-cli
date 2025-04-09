# list

List all partner ports

## Description

List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports. You can filter the partner ports based on specific criteria.

### Optional Fields
  - `company-name`: Filter partner ports by company name
  - `connect-type`: Filter partner ports by connect type
  - `diversity-zone`: Filter partner ports by diversity zone
  - `location-id`: Filter partner ports by location ID
  - `product-name`: Filter partner ports by product name

### Important Notes
  - The list can be filtered by multiple criteria at once
  - Filtering is case-insensitive and partial matches are supported

### Example Usage

```sh
  list
  list --product-name "AWS Partner Port"
  list --connect-type "Dedicated Cloud Connection"
  list --company-name "Amazon Web Services"
  list --location-id 1
  list --diversity-zone "blue"
```

## Usage

```sh
megaport-cli partners list [flags]
```


## Parent Command

* [megaport-cli partners](megaport-cli_partners.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_partners_list_docs.md)

