# delete

Delete a port from your account

## Description

Delete a port from your account in the Megaport API.

This command allows you to delete an existing port by providing the UID of the port as an argument.
By default, the port will be scheduled for deletion at the end of the current billing period.

Available flags:
  --now    Delete the port immediately instead of waiting until the end of the billing period.
           Note that immediate deletion may affect billing and cannot be undone.
           
  --force, -f  Skip the confirmation prompt and proceed with deletion.
               Use with caution, as this will immediately execute the delete operation.

Important notes:
- All VXCs associated with the port must be deleted before the port can be deleted
- You can restore a deleted port before it's fully decommissioned using the 'restore' command
- Once a port is fully decommissioned, restoration is not possible

Example usage:

### Delete at the end of the billing period (with confirmation prompt)
```
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

```

### Delete immediately (with confirmation prompt)
```
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now

```

### Delete immediately without confirmation
```
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now --force

```

Example output:
```
  Are you sure you want to delete port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p? (y/n): y
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p deleted successfully
  The port will be deleted at the end of the current billing period

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



