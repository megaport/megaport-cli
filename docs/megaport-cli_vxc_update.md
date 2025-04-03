# update

Update an existing Virtual Cross Connect (VXC)

## Description

Update an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to update an existing VXC by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode:
   The command will prompt you for each field that can be updated.

2. Flag Mode:
   Provide fields to update using flags:
   --name, --rate-limit, --term, --cost-centre, --shutdown, 
   --a-end-vlan, --b-end-vlan, --a-end-inner-vlan, --b-end-inner-vlan,
   --a-end-uid, --b-end-uid, --a-end-partner-config, --b-end-partner-config

3. JSON Mode:
   Provide a JSON string or file with update fields:
   --json <json-string> or --json-file <path>

Updateable fields:
- `name`: New name for the VXC
- `rate-limit`: New bandwidth in Mbps
- `term`: New contract term in months (1, 12, 24, or 36)
- `cost-centre`: New cost centre for billing
- `shutdown`: Whether to shut down the VXC (true/false)
- `a-end-vlan`: New VLAN for A-End (0-4093, except 1)
- `b-end-vlan`: New VLAN for B-End (0-4093, except 1)
- `a-end-inner-vlan`: New inner VLAN for A-End (-1 or higher)
- `b-end-inner-vlan`: New inner VLAN for B-End (-1 or higher)
- `a-end-uid`: New A-End product UID
- `b-end-uid`: New B-End product UID

NOTE: For partner configurations, only VRouter partner configurations can be updated.
Other CSP partner configurations (AWS, Azure, etc.) cannot be changed after creation.

Example usage:

# Interactive mode
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive
```

# Flag mode - Basic updates
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --name "New VXC Name" \
  --rate-limit 2000 \
  --cost-centre "New Cost Centre"
```

# Flag mode - Update VLANs
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --a-end-vlan 200 \
  --b-end-vlan 300
```

# Flag mode - Update with VRouter partner config
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --b-end-partner-config '{
    "interfaces": [
      {
        "vlan": 100,
        "ipAddresses": ["192.168.1.1/30"],
        "bgpConnections": [
          {
            "peerAsn": 65000,
            "localAsn": 64512,
            "localIpAddress": "192.168.1.1",
            "peerIpAddress": "192.168.1.2",
            "password": "bgppassword",
            "shutdown": false,
            "bfdEnabled": true
          }
        ]
      }
    ]
  }'
```

# JSON mode
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{
  "name": "Updated VXC Name",
  "rateLimit": 2000,
  "costCentre": "New Cost Centre",
  "aEndVlan": 200,
  "bEndVlan": 300,
  "term": 24,
  "shutdown": false
}'
```

# JSON file
```
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json-file ./vxc-update.json
```



## Usage

```
megaport-cli vxc update [vxcUID] [flags]
```



## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --a-end-inner-vlan |  | 0 | New inner VLAN for A-End | false |
| --a-end-partner-config |  |  | JSON string with A-End VRouter partner configuration | false |
| --a-end-uid |  |  | New A-End product UID | false |
| --a-end-vlan |  | 0 | New VLAN for A-End (0-4093, except 1) | false |
| --b-end-inner-vlan |  | 0 | New inner VLAN for B-End | false |
| --b-end-partner-config |  |  | JSON string with B-End VRouter partner configuration | false |
| --b-end-uid |  |  | New B-End product UID | false |
| --b-end-vlan |  | 0 | New VLAN for B-End (0-4093, except 1) | false |
| --cost-centre |  |  | New cost centre for billing | false |
| --interactive |  | false | Use interactive mode | false |
| --json |  |  | JSON string with update fields | false |
| --json-file |  |  | Path to JSON file with update fields | false |
| --name |  |  | New name for the VXC | false |
| --rate-limit |  | 0 | New bandwidth in Mbps | false |
| --shutdown |  | false | Whether to shut down the VXC | false |
| --term |  | 0 | New contract term in months (1, 12, 24, or 36) | false |



