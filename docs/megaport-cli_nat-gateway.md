# nat-gateway

Manage NAT Gateways in the Megaport API

## Description

Manage NAT Gateways in the Megaport API.

This command groups all operations related to Megaport NAT Gateways. NAT Gateways provide network address translation services within the Megaport fabric.

### Example Usage

```sh
  megaport-cli nat-gateway get [uid]
  megaport-cli nat-gateway list
  megaport-cli nat-gateway create --interactive
  megaport-cli nat-gateway update [uid]
  megaport-cli nat-gateway delete [uid]
  megaport-cli nat-gateway validate [uid]
  megaport-cli nat-gateway buy [uid]
  megaport-cli nat-gateway list-sessions
  megaport-cli nat-gateway telemetry [uid] --types BITS --days 7
```

## Usage

```sh
megaport-cli nat-gateway [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [buy](megaport-cli_nat-gateway_buy.md)
* [create](megaport-cli_nat-gateway_create.md)
* [delete](megaport-cli_nat-gateway_delete.md)
* [docs](megaport-cli_nat-gateway_docs.md)
* [get](megaport-cli_nat-gateway_get.md)
* [list](megaport-cli_nat-gateway_list.md)
* [list-sessions](megaport-cli_nat-gateway_list-sessions.md)
* [telemetry](megaport-cli_nat-gateway_telemetry.md)
* [update](megaport-cli_nat-gateway_update.md)
* [validate](megaport-cli_nat-gateway_validate.md)

