# ports

Manage ports in the Megaport API

## Description

Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands 
to list all ports, get details for a specific port, buy a new port, buy a LAG port,
update an existing port, delete a port, restore a deleted port, lock a port, unlock a port,
and check VLAN availability on a port.

Examples:
  # List all ports
  megaport-cli ports list

  # Get details for a specific port
  megaport-cli ports get [portUID]

  # Buy a new port
  megaport-cli ports buy --interactive
  megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
  megaport-cli ports buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
  megaport-cli ports buy --json-file ./port-config.json

  # Buy a LAG port
  megaport-cli ports buy-lag --interactive
  megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
  megaport-cli ports buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
  megaport-cli ports buy-lag --json-file ./lag-port-config.json

  # Update a port
  megaport-cli ports update [portUID] --interactive
  megaport-cli ports update [portUID] --name "Updated Port" --marketplace-visibility true
  megaport-cli ports update [portUID] --json '{"name":"Updated Port","marketplaceVisibility":true}'
  megaport-cli ports update [portUID] --json-file ./update-port-config.json

  # Delete a port
  megaport-cli ports delete [portUID] --now

  # Restore a deleted port
  megaport-cli ports restore [portUID]

  # Lock a port
  megaport-cli ports lock [portUID]

  # Unlock a port
  megaport-cli ports unlock [portUID]

  # Check VLAN availability on a port
  megaport-cli ports check-vlan [portUID] [vlan]



## Usage

```
megaport-cli ports [flags]
```









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

