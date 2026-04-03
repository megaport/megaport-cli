# delete

Delete an existing MVE

## Description

Delete an existing Megaport Virtual Edge (MVE).

This command allows you to delete an existing MVE by providing its UID.

### Important Notes
  - Deletion is final and cannot be undone
  - Billing for the MVE stops at the end of the current billing period unless --now is specified
  - All associated VXCs will be automatically terminated

### Example Usage

```sh
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --force
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now
```

## Usage

```sh
megaport-cli mve delete [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete resource immediately instead of at end of billing cycle | false |

## Subcommands
* [docs](megaport-cli_mve_delete_docs.md)

