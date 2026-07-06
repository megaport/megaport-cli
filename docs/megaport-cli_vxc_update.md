# update

Update an existing Virtual Cross Connect (VXC)

## Description

Update an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to update an existing VXC by providing the necessary details.

### Optional Fields
  - `a-end-inner-vlan`: New inner VLAN for A-End (0=none, -1=untagged, 2-4094 for specific VLAN (1 is reserved), only for QinQ)
  - `a-end-partner-config`: JSON string with A-End VRouter partner configuration
  - `a-end-uid`: New A-End product UID
  - `a-end-vlan`: New VLAN for A-End (0=auto-assign, -1=untagged, 2-4094 for specific VLAN (1 is reserved))
  - `b-end-inner-vlan`: New inner VLAN for B-End (0=none, -1=untagged, 2-4094 for specific VLAN (1 is reserved), only for QinQ)
  - `b-end-partner-config`: JSON string with B-End VRouter partner configuration
  - `b-end-uid`: New B-End product UID
  - `b-end-vlan`: New VLAN for B-End (0=auto-assign, -1=untagged, 2-4094 for specific VLAN (1 is reserved))
  - `cost-centre`: New cost centre for billing
  - `name`: New name for the VXC (1-64 characters)
  - `rate-limit`: New bandwidth in Mbps (50 - 10000)
  - `shutdown`: Whether to shut down the VXC (true/false)
  - `term`: New contract term in months (1, 12, 24, or 36)

### Important Notes
  - Only VRouter partner configurations can be updated after creation
  - CSP partner configurations (AWS, Azure, etc.) cannot be changed after creation
  - IPsec tunnels require interfaceType "ipSecTunnel" on the interface; treat the preSharedKey as a secret and avoid committing populated config files to source control
  - Changing the rate limit may result in additional charges
  - Updating VLANs will cause temporary disruption to the VXC connectivity

### Example Usage

```sh
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --name "New VXC Name" --rate-limit 2000 --cost-centre "New Cost Centre"
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --a-end-vlan 200 --b-end-vlan 300
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --b-end-partner-config '{"connectType":"VROUTER","interfaces":[{"vlan":100,"ipAddresses":["192.168.1.1/30"],"bgpConnections":[{"peerAsn":65000,"localAsn":64512,"localIpAddress":"192.168.1.1","peerIpAddress":"192.168.1.2","password":"<your-bgp-password>","shutdown":false,"bfdEnabled":true}]}]}'
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --a-end-partner-config '{"connectType":"VROUTER","interfaces":[{"interfaceType":"ipSecTunnel","ipSecTunnelOptions":{"sourceIpAddress":"192.0.2.1","destinationIpAddress":"198.51.100.1","preSharedKey":"<your-psk>","phase1Lifetime":28800,"phase2Lifetime":3600}}]}'
  megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{"name":"Updated VXC Name","rateLimit":2000,"costCentre":"New Cost Centre","aEndVlan":200,"bEndVlan":300,"term":24,"shutdown":false}'
```
### JSON Format Example
```json
{
  "name": "Updated VXC Name",
  "rateLimit": 2000,
  "costCentre": "New Cost Centre",
  "aEndVlan": 200,
  "bEndVlan": 300,
  "term": 24,
  "shutdown": false
}

```

## Usage

```sh
megaport-cli vxc update [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--a-end-inner-vlan` |  | `0` | Inner VLAN for A-End (0=none, -1=untagged, 2-4094 for specific VLAN (1 is reserved)) | false |
| `--a-end-partner-config` |  |  | JSON string with A-End partner configuration | false |
| `--a-end-uid` |  |  | UID of the A-End product | false |
| `--a-end-vlan` |  | `0` | VLAN for A-End (0=auto-assign, -1=untagged, 2-4094 for specific VLAN (1 is reserved)) | false |
| `--a-vnic-index` |  | `-1` | New A-End vNIC index when moving a VXC on an MVE | false |
| `--b-end-inner-vlan` |  | `0` | Inner VLAN for B-End (0=none, -1=untagged, 2-4094 for specific VLAN (1 is reserved)) | false |
| `--b-end-partner-config` |  |  | JSON string with B-End partner configuration | false |
| `--b-end-uid` |  |  | UID of the B-End product | false |
| `--b-end-vlan` |  | `0` | VLAN for B-End (0=auto-assign, -1=untagged, 2-4094 for specific VLAN (1 is reserved)) | false |
| `--b-vnic-index` |  | `-1` | New B-End vNIC index when moving a VXC on an MVE | false |
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--generate-skeleton` |  | `false` | Print a JSON skeleton template for --json or --json-file input and exit | false |
| `--interactive` |  | `false` | Use interactive mode | false |
| `--is-approved` |  | `false` | Approve or reject a VXC via the Megaport Marketplace | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--name` |  |  | Name of the VXC | false |
| `--rate-limit` |  | `0` | Bandwidth in Mbps | false |
| `--shutdown` |  | `false` | Whether to shut down the VXC | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | false |

