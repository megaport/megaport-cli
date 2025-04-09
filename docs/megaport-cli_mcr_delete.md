# delete

Delete an MCR from your account

## Description

Delete an MCR from your account.

This command allows you to delete an MCR from your account. By default, the MCR will be scheduled for deletion at the end of the current billing period.

### Example Usage

```sh
  delete [mcrUID]
  delete [mcrUID] --now
  delete [mcrUID] --now --force
```


## Usage

```sh
megaport-cli mcr delete [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete resource immediately instead of at end of billing cycle | false |



