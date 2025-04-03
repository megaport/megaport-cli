# mve

Manage Megaport Virtual Edge (MVE) devices

## Description

Manage Megaport Virtual Edge (MVE) devices.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
You can use this command to list, get details, buy, update, and delete MVEs.

Examples:
  # List all MVEs
  megaport-cli mve list

  # Get details for a specific MVE
  megaport-cli mve get [mveUID]

  # Buy a new MVE
  megaport-cli mve buy

  # Update an existing MVE
  megaport-cli mve update [mveUID]

  # Delete an existing MVE
  megaport-cli mve delete [mveUID]



## Usage

```
megaport-cli mve [flags]
```









## Subcommands

* [buy](mve_buy.md)
* [delete](mve_delete.md)
* [get](mve_get.md)
* [list-images](mve_list-images.md)
* [list-sizes](mve_list-sizes.md)
* [update](mve_update.md)

