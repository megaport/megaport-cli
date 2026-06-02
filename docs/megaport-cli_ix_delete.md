# delete

Delete an IX from your account

## Description

Delete an IX from your account.

Deletion is immediate by default; pass --later to schedule cancellation at the end of the current billing cycle instead.

### Important Notes
  - Deletion is immediate by default; billing stops right away
  - Use --later to defer cancellation to the end of the current billing cycle

### Example Usage

```sh
  megaport-cli ix delete [ixUID]
  megaport-cli ix delete [ixUID] --force
  megaport-cli ix delete [ixUID] --later
```

## Usage

```sh
megaport-cli ix delete [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)

## Aliases

* rm
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--later` |  | `false` | Schedule deletion at the end of the current billing cycle (default: delete immediately) | false |

## Subcommands
* [docs](megaport-cli_ix_delete_docs.md)

