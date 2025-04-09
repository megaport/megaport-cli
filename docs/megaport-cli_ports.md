# ports

Manage ports in the Megaport API

## Description

Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands to list all ports, get details for a specific port, buy a new port, buy a LAG port, update an existing port, delete a port, restore a deleted port, lock a port, unlock a port, and check VLAN availability on a port.

### Example Usage

```sh
  megaport-cli ports list
  megaport-cli ports get [portUID]
  megaport-cli ports buy --interactive
  megaport-cli ports buy-lag --interactive
  megaport-cli ports update [portUID] --name "Updated Port Name"
  megaport-cli ports delete [portUID]
  megaport-cli ports restore [portUID]
  megaport-cli ports lock [portUID]
  megaport-cli ports unlock [portUID]
  megaport-cli ports check-vlan [portUID] [vlan]
```

## Usage

```sh
megaport-cli ports [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [buy](megaport-cli_ports_buy.md)
* [buy-lag](megaport-cli_ports_buy-lag.md)
* [check-vlan](megaport-cli_ports_check-vlan.md)
* [delete](megaport-cli_ports_delete.md)
* [docs](megaport-cli_ports_docs.md)
* [get](megaport-cli_ports_get.md)
* [list](megaport-cli_ports_list.md)
* [lock](megaport-cli_ports_lock.md)
* [restore](megaport-cli_ports_restore.md)
* [unlock](megaport-cli_ports_unlock.md)
* [update](megaport-cli_ports_update.md)

