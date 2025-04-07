# ports

Manage ports in the Megaport API

## Description

Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands to list all ports, get details for a specific port, buy a new port, buy a LAG port, update an existing port, delete a port, restore a deleted port, lock a port, unlock a port, and check VLAN availability on a port.

Example usage:

```
  ports list
  ports get [portUID]
  ports buy --interactive
  ports buy-lag --interactive
  ports update [portUID] --name "Updated Port Name"
  ports delete [portUID]
  ports restore [portUID]
  ports lock [portUID]
  ports unlock [portUID]
  ports check-vlan [portUID] [vlan]
```


## Usage

```
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
* [get](megaport-cli_ports_get.md)
* [list](megaport-cli_ports_list.md)
* [lock](megaport-cli_ports_lock.md)
* [restore](megaport-cli_ports_restore.md)
* [unlock](megaport-cli_ports_unlock.md)
* [update](megaport-cli_ports_update.md)

