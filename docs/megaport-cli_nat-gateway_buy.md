# buy

Purchase a NAT Gateway design to begin provisioning

## Description

Purchase a NAT Gateway design via /v3/networkdesign/buy.

Run this after 'nat-gateway validate' to kick off provisioning of a NAT Gateway that currently exists in DESIGN state. Billing begins once the order is accepted.

### Important Notes
  - This action begins billing for the NAT Gateway and cannot be undone without deleting the resource.

### Example Usage

```sh
  megaport-cli nat-gateway buy a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli nat-gateway buy a1b2c3d4-e5f6-7890-1234-567890abcdef --yes
```

## Usage

```sh
megaport-cli nat-gateway buy [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--yes` |  | `false` | Skip the confirmation prompt | false |

