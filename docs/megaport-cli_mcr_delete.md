# delete

Delete an MCR from your account

## Description

Delete an MCR from your account.

MCRs are deleted immediately; deferred end-of-term cancellation is not supported by the API.

### Important Notes
  - Deletion is immediate; billing stops right away
  - Deletion is final and cannot be undone

### Example Usage

```sh
  megaport-cli mcr delete [mcrUID]
  megaport-cli mcr delete [mcrUID] --force
  megaport-cli mcr delete [mcrUID] --force --safe-delete
```

## Usage

```sh
megaport-cli mcr delete [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)

## Aliases

* rm
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--safe-delete` |  | `false` | Fail if the resource has attached VXCs or other active services | false |

