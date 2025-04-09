# vxc

Manage VXCs in the Megaport API

## Description

Manage VXCs in the Megaport API.

This command groups all operations related to Virtual Cross Connects (VXCs). VXCs are virtual point-to-point connections between two ports or devices on the Megaport network. You can use the subcommands to perform actions such as retrieving details, purchasing, updating, and deleting VXCs.

### Example Usage

```sh
  megaport-cli vxc get [vxcUID]
  megaport-cli vxc buy
  megaport-cli vxc update [vxcUID]
  megaport-cli vxc delete [vxcUID]
```

## Usage

```sh
megaport-cli vxc [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [buy](megaport-cli_vxc_buy.md)
* [delete](megaport-cli_vxc_delete.md)
* [docs](megaport-cli_vxc_docs.md)
* [get](megaport-cli_vxc_get.md)
* [update](megaport-cli_vxc_update.md)

