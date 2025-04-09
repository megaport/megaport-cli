# servicekeys

Manage service keys for the Megaport API

## Description

Manage service keys for the Megaport API.

This command groups all operations related to service keys. You can use its subcommands to create, update, list, and get details of service keys.

### Example Usage

```sh
  servicekeys list
  servicekeys get [key]
  servicekeys create --product-uid "product-uid" --description "My service key"
  servicekeys update [key] --description "Updated description"
```

## Usage

```sh
megaport-cli servicekeys [flags]
```




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|


## Subcommands

* [create](megaport-cli_servicekeys_create.md)
* [get](megaport-cli_servicekeys_get.md)
* [list](megaport-cli_servicekeys_list.md)
* [update](megaport-cli_servicekeys_update.md)

