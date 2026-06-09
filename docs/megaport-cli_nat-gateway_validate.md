# validate

Validate a NAT Gateway design without purchasing

## Description

Validate a NAT Gateway design via /v3/networkdesign/validate.

Use this after 'nat-gateway create' to preview pricing and confirm the design is valid before calling 'nat-gateway buy'. No resources are provisioned and no charges are incurred.

### Important Notes
  - The NAT Gateway must already exist in DESIGN state (create it first with 'nat-gateway create').

### Example Usage

```sh
  megaport-cli nat-gateway validate a1b2c3d4-e5f6-7890-1234-567890abcdef
  megaport-cli nat-gateway validate a1b2c3d4-e5f6-7890-1234-567890abcdef --output json
```

## Usage

```sh
megaport-cli nat-gateway validate [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_nat-gateway_validate_docs.md)

