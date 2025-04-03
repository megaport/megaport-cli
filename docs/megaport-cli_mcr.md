# mcr

Manage MCRs in the Megaport API

## Description

Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
You can use the subcommands to perform actions such as retrieving details for a specific MCR.
For instance, use the "megaport-cli mcr get [mcrUID]" command to fetch details for the MCR with the given UID.

Available subcommands:
- `get`: Retrieve details for a single MCR.
- `buy`: Purchase an MCR by providing the necessary details.
- `update`: Update an existing MCR.
- `delete`: Delete an MCR from your account.
- `restore`: Restore a previously deleted MCR.
- `create-prefix-filter-list`: Create a prefix filter list on an MCR.
- `list-prefix-filter-lists`: List all prefix filter lists for a specific MCR.
- `get-prefix-filter-list`: Retrieve details for a single prefix filter list on an MCR.
- `update-prefix-filter-list`: Update a prefix filter list on an MCR.
- `delete-prefix-filter-list`: Delete a prefix filter list on an MCR.



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

