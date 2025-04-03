# buy

Purchase a new Virtual Cross Connect (VXC)

## Description

Purchase a new Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to purchase a VXC by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (default):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --a-end-uid, --b-end-uid, --name, --rate-limit, --term,
   --a-end-vlan, --b-end-vlan, --a-end-partner-config, --b-end-partner-config

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
- `a-end-uid`: UID of the A-End product
- `name`: Name of the VXC
- `rate-limit`: Bandwidth in Mbps
- `term`: Contract term in months (1, 12, 24, or 36)

Optional fields:
- `b-end-uid`: UID of the B-End product (if connecting to non-partner)
- `a-end-vlan`: VLAN for A-End (0-4093, except 1)
- `b-end-vlan`: VLAN for B-End (0-4093, except 1)
- `a-end-inner-vlan`: Inner VLAN for A-End (-1 or higher)
- `b-end-inner-vlan`: Inner VLAN for B-End (-1 or higher)
- `a-end-vnic-index`: vNIC index for A-End MVE
- `b-end-vnic-index`: vNIC index for B-End MVE
- `promo-code`: Promotional code
- `service-key`: Service key
- `cost-centre`: Cost centre
- `a-end-partner-config`: JSON string with A-End partner configuration
- `b-end-partner-config`: JSON string with B-End partner configuration

Example usage:

### Interactive mode
```
megaport-cli vxc buy --interactive

```

### Flag mode - Basic VXC between two ports
```
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-vlan 200

```

### Flag mode - VXC to AWS Direct Connect
```
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My AWS VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
- `-b-end-partner-config '{"connectType"`: "AWS","ownerAccount":"123456789012","asn":65000,"amazonAsn":64512}'

```

### Flag mode - VXC to Azure ExpressRoute
```
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --name "My Azure VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
- `-b-end-partner-config '{"connectType"`: "AZURE","serviceKey":"s-abcd1234"}'

```

### JSON mode
```
megaport-cli vxc buy --json '{
  "portUID": "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "vxcName": "My VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {
    "vlan": 100
  },
  "bEndConfiguration": {
    "productUid": "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "vlan": 200
  }
}'

```

### JSON mode with partner config
```
megaport-cli vxc buy --json '{
  "portUID": "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "vxcName": "My AWS VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {
    "vlan": 100
  },
  "bEndConfiguration": {
    "productUid": "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "partnerConfig": {
      "connectType": "AWS",
      "ownerAccount": "123456789012",
      "asn": 65000,
      "amazonAsn": 64512,
      "type": "private"
    }
  }
}'

```

### JSON file
```
megaport-cli vxc buy --json-file ./vxc-config.json

```



## Usage

```
megaport-cli vxc buy [flags]
```



## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --a-end-inner-vlan |  | 0 | Inner VLAN for A-End (-1 or higher) | false |
| --a-end-partner-config |  |  | JSON string with A-End partner configuration | false |
| --a-end-uid |  |  | UID of the A-End product | false |
| --a-end-vlan |  | 0 | VLAN for A-End (0-4093, except 1) | false |
| --a-end-vnic-index |  | 0 | vNIC index for A-End MVE | false |
| --b-end-inner-vlan |  | 0 | Inner VLAN for B-End (-1 or higher) | false |
| --b-end-partner-config |  |  | JSON string with B-End partner configuration | false |
| --b-end-uid |  |  | UID of the B-End product | false |
| --b-end-vlan |  | 0 | VLAN for B-End (0-4093, except 1) | false |
| --b-end-vnic-index |  | 0 | vNIC index for B-End MVE | false |
| --cost-centre |  |  | Cost centre | false |
| --interactive |  | false | Use interactive mode | false |
| --json |  |  | JSON string with all VXC configuration | false |
| --json-file |  |  | Path to JSON file with VXC configuration | false |
| --name |  |  | Name of the VXC | false |
| --promo-code |  |  | Promotional code | false |
| --rate-limit |  | 0 | Bandwidth in Mbps | false |
| --service-key |  |  | Service key | false |
| --term |  | 0 | Contract term in months (1, 12, 24, or 36) | false |



