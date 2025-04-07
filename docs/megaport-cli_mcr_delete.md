# delete

Delete an MCR from your account

## Description

Delete an MCR from your account.

This command allows you to delete an MCR from your account. By default, the MCR will be scheduled for deletion at the end of the current billing period.

Optional fields:
  force: Skip the confirmation prompt and proceed with deletion
  now: Delete the MCR immediately instead of at the end of the billing period

Example usage:

```
  delete [mcrUID]
  delete [mcrUID] --now
  delete [mcrUID] --now --force
```


## Usage

```
megaport-cli mcr delete [mcrUID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Force deletion without confirmation | false |
| `--now` |  | `false` | Delete MCR immediately instead of at end of billing cycle | false |



