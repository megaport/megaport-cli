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

```
  delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```


## Usage

```
megaport-cli vxc delete [vxcUID] [flags]
```



## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|



