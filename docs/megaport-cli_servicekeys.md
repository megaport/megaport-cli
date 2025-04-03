# servicekeys

Manage service keys for the Megaport API

## Description

Manage service keys for the Megaport API.

This command groups all operations related to service keys. You can use its subcommands to:
  - Create a new service key.
  - Update an existing service key.
  - List all service keys.
  - Get details of a specific service key.

Examples:
  megaport-cli servicekeys list
  megaport-cli servicekeys get [key]
  megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"
  megaport-cli servicekeys update [key] --description "Updated description"



## Usage

```
megaport-cli servicekeys [flags]
```









## Subcommands

* [create](servicekeys_create.md)
* [get](servicekeys_get.md)
* [list](servicekeys_list.md)
* [update](servicekeys_update.md)

