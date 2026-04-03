# get

Get details for a single IX

## Description

Get details for a single IX.

This command retrieves and displays detailed information for a single Internet Exchange (IX). You must provide the unique identifier (UID) of the IX you wish to retrieve.

### Important Notes
  - The output includes the IX's UID, name, network service type, ASN, rate limit, VLAN, MAC address, and provisioning status

### Example Usage

```sh
  megaport-cli ix get a1b2c3d4-e5f6-7890-1234-567890abcdef
```

## Usage

```sh
megaport-cli ix get [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_ix_get_docs.md)

