# check-vlan

Check if a VLAN is available on a port

## Description

Check if a VLAN is available on a port in the Megaport API.

This command verifies whether a specific VLAN ID is available for use on a port. This is useful when planning new VXCs to ensure the VLAN ID you want to use is not already in use by another connection.

VLAN ID must be between 2 and 4094 (inclusive).

### Example Usage

```sh
  megaport-cli ports check-vlan 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p 100
  megaport-cli ports check-vlan port-abc123 500
```

## Usage

```sh
megaport-cli ports check-vlan [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

