# unlock

Unlock a port

## Description

Unlock a port in the Megaport API.

This command allows you to unlock a previously locked port, re-enabling the ability
to make changes to the port and its associated VXCs.

When a port is unlocked:
- The port's configuration can be modified
- New VXCs can be created on this port
- Existing VXCs can be modified or deleted
- The port itself can be deleted if needed

Example usage:

```
  megaport-cli ports unlock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

```

Example output:
```
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p unlocked successfully

```



## Usage

```
megaport-cli ports unlock [portUID] [flags]
```



## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)







