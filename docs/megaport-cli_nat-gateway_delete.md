# delete

Delete a NAT Gateway

## Description

Delete a NAT Gateway.

This command deletes an existing NAT Gateway by its product UID.

### Important Notes
  - This action is irreversible. The NAT Gateway will be deleted immediately.

### Example Usage

```sh
  megaport-cli nat-gateway delete a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli nat-gateway delete a1b2c3d4-e5f6-7890-1234-567890abcdef --force
```

## Usage

```sh
megaport-cli nat-gateway delete [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)

## Aliases

* rm
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Skip the confirmation prompt | false |
| `--yes` |  | `false` | Skip the confirmation prompt (alias of --force) | false |

