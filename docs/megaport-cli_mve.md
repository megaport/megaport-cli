# mve

Manage Megaport Virtual Edge (MVE) devices

## Description

Manage Megaport Virtual Edge (MVE) devices.

This command groups all operations related to Megaport Virtual Edge devices (MVEs). MVEs are virtual networking appliances that run in the Megaport network, providing software-defined networking capabilities from various vendors.

### Important Notes
  - With MVEs you can deploy virtual networking appliances without physical hardware
  - Create secure connections between cloud services
  - Run vendor-specific networking software in Megaport's infrastructure

### Example Usage

```sh
  mve list
  mve get [mveUID]
  mve buy
  mve update [mveUID]
  mve delete [mveUID]
```

## Usage

```sh
megaport-cli mve [flags]
```




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|


## Subcommands

* [buy](megaport-cli_mve_buy.md)
* [delete](megaport-cli_mve_delete.md)
* [get](megaport-cli_mve_get.md)
* [list-images](megaport-cli_mve_list-images.md)
* [list-sizes](megaport-cli_mve_list-sizes.md)
* [update](megaport-cli_mve_update.md)

