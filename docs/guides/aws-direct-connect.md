# Connect to AWS Direct Connect

This guide walks through creating an AWS Direct Connect connection using the Megaport CLI — from finding a location to verifying the live connection.

## Prerequisites

- Megaport account with API credentials ([Megaport Portal](https://portal.megaport.com))
- AWS account with Direct Connect permissions
- `megaport-cli` installed ([installation instructions](../../README.md#installation))

## 1. Configure the CLI

Create a named profile with your Megaport API credentials:

```sh
megaport-cli config create-profile aws-guide \
  --access-key YOUR_ACCESS_KEY \
  --secret-key YOUR_SECRET_KEY \
  --environment production

megaport-cli config use-profile aws-guide
```

Verify your credentials are working:

```sh
megaport-cli locations list-countries
```

## 2. Find a Location

Search for a data centre near you:

```sh
megaport-cli locations search "Sydney"
```

Or filter by country:

```sh
megaport-cli locations list --country "AU"
```

Note the **Location ID** from the output — you'll use it in the next steps.

## 3. Create a Port

Buy a port at your chosen location. Use `--marketplace-visibility false` if you don't want the port listed publicly:

```sh
megaport-cli ports buy \
  --name "AWS-Port-SYD" \
  --port-speed 10000 \
  --location-id 15 \
  --term 12 \
  --marketplace-visibility false
```

The output includes your new **Port UID** (e.g. `port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`). Use `--no-wait` to skip waiting for the port to reach LIVE status.

Check provisioning status:

```sh
megaport-cli ports status port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## 4. Find the AWS Partner Port

List Amazon Web Services partner ports at your location:

```sh
megaport-cli partners list \
  --company-name "Amazon Web Services" \
  --location-id 15
```

Note the **Product UID** of the AWS partner port — this is your B-End for the VXC.

## 5. Create a VXC to AWS

Create the Virtual Cross Connect between your port and the AWS partner port. The `--b-end-partner-config` flag passes your AWS account details:

```sh
megaport-cli vxc buy \
  --name "AWS-DX-SYD" \
  --rate-limit 500 \
  --term 12 \
  --a-end-uid port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AWS","ownerAccount":"123456789012"}'
```

For a connection with full BGP configuration:

```sh
megaport-cli vxc buy \
  --name "AWS-DX-SYD" \
  --rate-limit 500 \
  --term 12 \
  --a-end-uid port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --a-end-vlan 100 \
  --b-end-partner-config '{
    "connectType": "AWS",
    "ownerAccount": "123456789012",
    "asn": 65000,
    "amazonAsn": 64512,
    "authKey": "your-bgp-auth-key",
    "customerIPAddress": "169.254.10.1/30",
    "amazonIPAddress": "169.254.10.2/30"
  }'
```

> **Hosted Connection:** Use `"connectType": "AWSHC"` for an AWS Hosted Connection (lower-bandwidth, shared capacity).

The output includes your new **VXC UID**.

### Partner config fields

| Field | Required | Description |
|---|---|---|
| `connectType` | Yes | `"AWS"` (dedicated) or `"AWSHC"` (hosted) |
| `ownerAccount` | Yes | Your 12-digit AWS account ID |
| `asn` | No | Your BGP ASN |
| `amazonAsn` | No | Amazon's BGP ASN |
| `authKey` | No | BGP authentication key |
| `customerIPAddress` | No | Your BGP peer IP (CIDR, e.g. `169.254.10.1/30`) |
| `amazonIPAddress` | No | Amazon's BGP peer IP (CIDR, e.g. `169.254.10.2/30`) |

## 6. Verify the Connection

Check that the VXC has reached LIVE status:

```sh
megaport-cli vxc status vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

Get full connection details:

```sh
megaport-cli vxc get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## 7. Configure the AWS Side

Once the VXC is LIVE, complete the setup in the AWS Console:

1. Open **AWS Direct Connect** → **Virtual Interfaces**
2. Create a new Virtual Interface (Private or Transit) using the connection ID Megaport provides
3. Configure BGP using the ASN and IP addresses from your `--b-end-partner-config` values

## Teardown

To remove the connection when no longer needed:

```sh
# Delete VXC first
megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --now

# Then delete the port
megaport-cli ports delete port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --now
```

## Related Commands

- [`megaport-cli locations`](../megaport-cli_locations.md) — explore available locations
- [`megaport-cli partners`](../megaport-cli_partners.md) — find cloud provider partner ports
- [`megaport-cli ports`](../megaport-cli_ports.md) — manage physical ports
- [`megaport-cli vxc`](../megaport-cli_vxc.md) — manage Virtual Cross Connects
