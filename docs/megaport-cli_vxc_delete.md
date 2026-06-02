# delete

Delete an existing Virtual Cross Connect (VXC)

## Description

Delete an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to delete an existing VXC by providing its UID. Deletion is immediate by default; pass --later to schedule cancellation at the end of the current billing cycle instead.

### Important Notes
  - Deletion is immediate by default; the VXC is disconnected and billing stops right away
  - Use --later to defer cancellation to the end of the current billing cycle (not supported for Transit VXCs)
  - Deletion is final and cannot be undone

### Example Usage

```sh
  megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --force
  megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --later
```

## Usage

```sh
megaport-cli vxc delete [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)

## Aliases

* rm
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--later` |  | `false` | Schedule deletion at the end of the current billing cycle (default: delete immediately) | false |

## Subcommands
* [docs](megaport-cli_vxc_delete_docs.md)

