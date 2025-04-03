# delete

Delete a port from your account

## Description

Delete a port from your account in the Megaport API.

This command allows you to delete an existing port by providing the UID of the port as an argument.
You can optionally specify whether to delete the port immediately or at the end of the billing period.

Example usage:

```
  megaport-cli ports delete [portUID]

```



## Usage

```
megaport-cli ports delete [portUID] [flags]
```



## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete immediately instead of at the end of the billing period | false |



