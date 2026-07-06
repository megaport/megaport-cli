# ip-routes

List IP routes from the MCR routing table

## Description

List IP routes from the MCR Looking Glass.

This command retrieves all routes (BGP, static, connected, local) from the MCR's routing table. You can filter by protocol or IP address/prefix.

### Important Notes
  - Protocol values: BGP, STATIC, CONNECTED, LOCAL

### Example Usage

```sh
  megaport-cli mcr looking-glass ip-routes [mcrUID]
  megaport-cli mcr looking-glass ip-routes [mcrUID] --protocol BGP
  megaport-cli mcr looking-glass ip-routes [mcrUID] --ip 10.0.0.0/8
  megaport-cli mcr looking-glass ip-routes [mcrUID] --protocol STATIC --ip 192.168.0.0/16
```

## Usage

```sh
megaport-cli mcr looking-glass ip-routes [flags]
```


## Parent Command

* [megaport-cli mcr looking-glass](megaport-cli_mcr_looking-glass.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--ip` |  |  | Filter by IP address or prefix (e.g., 10.0.0.0/8 or 192.168.1.1) | false |
| `--protocol` |  |  | Filter by protocol (BGP, STATIC, CONNECTED, LOCAL) | false |

