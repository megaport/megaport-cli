# list

List all NAT Gateways

## Description

List all NAT Gateways for your account.

This command retrieves and displays all NAT Gateways, with optional filtering.

### Optional Fields
  - `include-inactive`: Include inactive NAT Gateways in the results
  - `limit`: Limit the number of results returned
  - `location-id`: Filter NAT Gateways by location ID
  - `name`: Filter NAT Gateways by name (substring match)

### Example Usage

```sh
  megaport-cli nat-gateway list
  megaport-cli nat-gateway list --location-id 67
  megaport-cli nat-gateway list --name "my-gw"
  megaport-cli nat-gateway list --include-inactive
```

## Usage

```sh
megaport-cli nat-gateway list [flags]
```


## Parent Command

* [megaport-cli nat-gateway](megaport-cli_nat-gateway.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive NAT Gateways in the list | false |
| `--limit` |  | `0` | Limit the number of results returned | false |
| `--location-id` |  | `0` | Filter NAT Gateways by location ID | false |
| `--name` |  |  | Filter NAT Gateways by name (substring match) | false |

## Subcommands
* [docs](megaport-cli_nat-gateway_list_docs.md)

