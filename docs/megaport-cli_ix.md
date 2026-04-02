# ix

Manage Internet Exchanges (IXs) in the Megaport API

## Description

Manage Internet Exchanges (IXs) in the Megaport API.

This command groups all operations related to Megaport Internet Exchange connections. IXs allow you to connect to Internet Exchange points through the Megaport fabric.

### Important Notes
  - IXs allow you to connect to Internet Exchange points for peering
  - An IX is attached to an existing port via the product-uid flag

### Example Usage

```sh
  megaport-cli ix get [ixUID]
  megaport-cli ix list
  megaport-cli ix buy
  megaport-cli ix update [ixUID]
  megaport-cli ix delete [ixUID]
```

## Usage

```sh
megaport-cli ix [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [buy](megaport-cli_ix_buy.md)
* [delete](megaport-cli_ix_delete.md)
* [docs](megaport-cli_ix_docs.md)
* [get](megaport-cli_ix_get.md)
* [list](megaport-cli_ix_list.md)
* [status](megaport-cli_ix_status.md)
* [update](megaport-cli_ix_update.md)
* [validate](megaport-cli_ix_validate.md)

