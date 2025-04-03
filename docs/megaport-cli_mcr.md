# mcr

Manage MCRs in the Megaport API

## Description

Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
MCRs are virtual routing appliances that run in the Megaport network, providing
interconnection between your cloud environments and the Megaport fabric.

With MCRs you can:
- Establish virtual cross-connects (VXCs) to cloud service providers
- Create private network connections between different cloud regions
- Implement hybrid cloud architectures with seamless connectivity
- Peer with other networks using BGP routing

Available operations:
- `get`: Retrieve details for a single MCR.
- `buy`: Purchase a new MCR with specified configuration.
- `update`: Modify an existing MCR's properties.
- `delete`: Remove an MCR from your account.
- `restore`: Restore a previously deleted MCR.
- `create-prefix-filter-list`: Create a prefix filter list on an MCR.
- `list-prefix-filter-lists`: List all prefix filter lists for a specific MCR.
- `get-prefix-filter-list`: Retrieve details for a single prefix filter list on an MCR.
- `update-prefix-filter-list`: Update a prefix filter list on an MCR.
- `delete-prefix-filter-list`: Delete a prefix filter list on an MCR.

Examples:
### Get details for a specific MCR
```
  megaport-cli mcr get [mcrUID]

```

### Buy a new MCR
```
  megaport-cli mcr buy

```

### Update an existing MCR
```
  megaport-cli mcr update [mcrUID]

```

### Delete an existing MCR
```
  megaport-cli mcr delete [mcrUID]

```



## Usage

```
megaport-cli mcr [flags]
```









## Subcommands

* [buy](megaport-cli_mcr_buy.md)
* [create-prefix-filter-list](megaport-cli_mcr_create-prefix-filter-list.md)
* [delete](megaport-cli_mcr_delete.md)
* [delete-prefix-filter-list](megaport-cli_mcr_delete-prefix-filter-list.md)
* [get](megaport-cli_mcr_get.md)
* [get-prefix-filter-list](megaport-cli_mcr_get-prefix-filter-list.md)
* [list-prefix-filter-lists](megaport-cli_mcr_list-prefix-filter-lists.md)
* [restore](megaport-cli_mcr_restore.md)
* [update](megaport-cli_mcr_update.md)
* [update-prefix-filter-list](megaport-cli_mcr_update-prefix-filter-list.md)

