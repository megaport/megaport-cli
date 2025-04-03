# mve

Manage Megaport Virtual Edge (MVE) devices

## Description

Manage Megaport Virtual Edge (MVE) devices.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
MVEs are virtual networking appliances that run in the Megaport network, providing
software-defined networking capabilities from various vendors.

With MVEs you can:
- Deploy virtual networking appliances without physical hardware
- Create secure connections between cloud services
- Implement SD-WAN solutions across multiple regions
- Run vendor-specific networking software in Megaport's infrastructure

Available operations:
- list: List all MVEs in your account
- get: Retrieve details for a specific MVE
- buy: Purchase a new MVE with vendor-specific configuration
- update: Modify an existing MVE's properties
- delete: Remove an MVE from your account
- list-images: View available MVE software images
- list-sizes: View available MVE hardware configurations

Examples:
### List all MVEs
```
megaport-cli mve list

```
### Get details for a specific MVE
```
megaport-cli mve get [mveUID]

```
### Buy a new MVE
```
megaport-cli mve buy

```
### Update an existing MVE
```
megaport-cli mve update [mveUID]

```
### Delete an existing MVE
```
megaport-cli mve delete [mveUID]

```


## Usage

```
megaport-cli mve [flags]
```







## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|


## Subcommands

* [buy](megaport-cli_mve_buy.md)
* [delete](megaport-cli_mve_delete.md)
* [get](megaport-cli_mve_get.md)
* [list](megaport-cli_mve_list.md)
* [list-images](megaport-cli_mve_list-images.md)
* [list-sizes](megaport-cli_mve_list-sizes.md)
* [update](megaport-cli_mve_update.md)

