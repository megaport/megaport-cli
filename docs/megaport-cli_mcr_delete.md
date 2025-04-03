# delete

Delete an MCR from your account

## Description

Delete an MCR from your account.

This command allows you to delete an MCR from your account. By default, the MCR
will be scheduled for deletion at the end of the current billing period.

Flags:
--now: Delete the MCR immediately instead of at the end of the billing period.
--force, -f: Skip the confirmation prompt and proceed with deletion.

Example usage:
### Delete MCR at the end of the billing period with confirmation
```
megaport-cli mcr delete [mcrUID]

```
### Delete MCR immediately with confirmation
```
megaport-cli mcr delete [mcrUID] --now

```
### Delete MCR immediately without confirmation
```
megaport-cli mcr delete [mcrUID] --now --force

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
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete immediately instead of at the end of the billing period | false |



