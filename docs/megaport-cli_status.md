# status

Show a dashboard of all Megaport resources

## Description

Display a combined status view of all Megaport resources.

Fetches ports, MCRs, MVEs, VXCs, and IXs in parallel and displays them in a single dashboard. By default, only active resources are shown.

### Example Usage

```sh
  megaport-cli status
  megaport-cli status --output json
  megaport-cli status --include-inactive
```

## Usage

```sh
megaport-cli status [flags]
```



## Aliases

* st
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive/decommissioned resources | false |

## Subcommands
* [docs](megaport-cli_status_docs.md)

