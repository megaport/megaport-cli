# delete

Delete an IX from your account

## Description

Delete an IX from your account.

This command allows you to delete an IX from your account. By default, the IX will be scheduled for deletion at the end of the current billing period.

### Example Usage

```sh
  megaport-cli ix delete [ixUID]
  megaport-cli ix delete [ixUID] --now
  megaport-cli ix delete [ixUID] --now --force
```

## Usage

```sh
megaport-cli ix delete [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete resource immediately instead of at end of billing cycle | false |

## Subcommands
* [docs](megaport-cli_ix_delete_docs.md)

