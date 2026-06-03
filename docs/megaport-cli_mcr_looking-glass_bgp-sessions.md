# bgp-sessions

List BGP sessions configured on the MCR

## Description

List all BGP sessions configured on the MCR.

This command shows the status of all BGP peering sessions including neighbor address, ASN, session status, uptime, and prefix counts.

### Important Notes
  - Session status can be UP, DOWN, or UNKNOWN
  - Use the session ID from this output with bgp-neighbor-routes command

### Example Usage

```sh
  megaport-cli mcr looking-glass bgp-sessions [mcrUID]
```

## Usage

```sh
megaport-cli mcr looking-glass bgp-sessions [flags]
```


## Parent Command

* [megaport-cli mcr looking-glass](megaport-cli_mcr_looking-glass.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_mcr_looking-glass_bgp-sessions_docs.md)

