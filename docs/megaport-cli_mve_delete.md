# delete

Delete an existing MVE

## Description

Delete an existing Megaport Virtual Edge (MVE).

MVEs are deleted immediately; deferred end-of-term cancellation is not supported by the API.

### Important Notes
  - Deletion is immediate; billing stops right away
  - All associated VXCs will be automatically terminated
  - Deletion is final and cannot be undone

### Example Usage

```sh
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --force
  megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --force --safe-delete
```

## Usage

```sh
megaport-cli mve delete [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)

## Aliases

* rm
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--safe-delete` |  | `false` | Fail if the resource has attached VXCs or other active services | false |

## Subcommands
* [docs](megaport-cli_mve_delete_docs.md)

