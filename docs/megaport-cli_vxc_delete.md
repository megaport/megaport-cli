# delete

Delete an existing Virtual Cross Connect (VXC)

## Description

Delete an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to delete an existing VXC by providing its UID.

### Important Notes
  - Deletion is final and cannot be undone
  - Billing for the VXC stops at the end of the current billing period
  - The VXC is immediately disconnected upon deletion

### Example Usage

```sh
  delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --force
  delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --now
```


## Usage

```sh
megaport-cli vxc delete [flags]
```



## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete resource immediately instead of at end of billing cycle | false |



