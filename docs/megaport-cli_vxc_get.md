# get

Get details for a single VXC

## Description

Get details for a single VXC through the Megaport API.

This command retrieves detailed information for a single Virtual Cross Connect (VXC). You must provide the unique identifier (UID) of the VXC you wish to retrieve.

### Important Notes
  - The output includes the VXC's UID, name, rate limit, A-End and B-End details, status, and cost centre.

### Example Usage

```sh
  megaport-cli vxc get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  megaport-cli vxc get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --export
```

## Usage

```sh
megaport-cli vxc get [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--export` |  | `false` | Output recreatable JSON config for use with buy --json (excludes read-only fields; partner configs not available from API) | false |

## Subcommands
* [docs](megaport-cli_vxc_get_docs.md)

