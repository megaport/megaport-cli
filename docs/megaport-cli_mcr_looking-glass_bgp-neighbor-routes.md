# bgp-neighbor-routes

List routes advertised to or received from a BGP neighbor

## Description

List routes advertised to or received from a specific BGP neighbor.

This command shows routes that are either being advertised to a neighbor or received from a neighbor. Use the session ID from 'bgp-sessions' command.

### Important Notes
  - Direction must be 'advertised' or 'received'
  - Get session ID from 'mcr looking-glass bgp-sessions' command

### Example Usage

```sh
  megaport-cli mcr looking-glass bgp-neighbor-routes [mcrUID] [sessionID] advertised
  megaport-cli mcr looking-glass bgp-neighbor-routes [mcrUID] [sessionID] received
  megaport-cli mcr looking-glass bgp-neighbor-routes [mcrUID] [sessionID] received --ip 10.0.0.0/8
```

## Usage

```sh
megaport-cli mcr looking-glass bgp-neighbor-routes [flags]
```


## Parent Command

* [megaport-cli mcr looking-glass](megaport-cli_mcr_looking-glass.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--ip` |  |  | Filter by IP address or prefix (e.g., 10.0.0.0/8 or 192.168.1.1) | false |

