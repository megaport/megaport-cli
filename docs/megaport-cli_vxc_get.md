# get

Get details for a single VXC

## Description

Get details for a single VXC through the Megaport API.

This command retrieves detailed information for a single Virtual Cross Connect (VXC).
You must provide the unique identifier (UID) of the VXC you wish to retrieve.

The output includes:
- `UID`: Unique identifier of the VXC
- `Name`: User-defined name of the VXC
- `Rate Limit`: Bandwidth of the VXC in Mbps
- `A-End`: Details of the A-End connection point
- `B-End`: Details of the B-End connection point
- `Status`: Current status of the VXC (e.g., Active, Inactive, Deleting)
- `Cost Centre`: Cost center associated with the VXC

Example usage:
```
  megaport-cli vxc get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

```

Example output:
```
  UID:          vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  Name:         My VXC
  Rate Limit:   1000 Mbps
  A-End:        Port: port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy, VLAN: 100
  B-End:        Port: port-zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz, VLAN: 200
  Status:       Active
  Cost Centre:  IT-Networking

```



## Usage

```
megaport-cli vxc get [vxcUID] [flags]
```



## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)







