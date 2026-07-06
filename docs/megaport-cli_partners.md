# partners

Manage partner ports in the Megaport API

## Description

Manage partner ports in the Megaport API.

This command groups all operations related to partner ports. You can use its subcommands to list and filter available partner ports based on specific criteria.

### Example Usage

```sh
  megaport-cli partners find
  megaport-cli partners list
  megaport-cli partners list --product-name "AWS Partner Port" --company-name "AWS" --location-id 1
```

## Usage

```sh
megaport-cli partners [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [find](megaport-cli_partners_find.md)
* [list](megaport-cli_partners_list.md)

