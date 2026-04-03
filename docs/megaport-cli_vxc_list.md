# list

List all VXCs with optional filters

## Description

List all VXCs available in the Megaport API.

This command retrieves all Virtual Cross Connects (VXCs) associated with your account. You can filter results by name, rate limit, A-End UID, B-End UID, or status.

### Optional Fields
  - `a-end-uid`: Filter VXCs by A-End product UID
  - `b-end-uid`: Filter VXCs by B-End product UID
  - `include-inactive`: Include inactive VXCs in the list
  - `name`: Filter VXCs by name (partial match)
  - `name-contains`: Filter VXCs by partial name match (server-side)
  - `rate-limit`: Filter VXCs by rate limit in Mbps
  - `status`: Filter VXCs by status (comma-separated, e.g. LIVE,CONFIGURED)

### Example Usage

```sh
  megaport-cli vxc list
  megaport-cli vxc list --name "My VXC"
  megaport-cli vxc list --a-end-uid port-abc123
  megaport-cli vxc list --status LIVE,CONFIGURED
  megaport-cli vxc list --include-inactive
```

## Usage

```sh
megaport-cli vxc list [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--a-end-uid` |  |  | Filter VXCs by A-End product UID | false |
| `--b-end-uid` |  |  | Filter VXCs by B-End product UID | false |
| `--include-inactive` |  | `false` | Include inactive VXCs in the list | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--name` |  |  | Filter VXCs by name (partial match) | false |
| `--name-contains` |  |  | Filter VXCs by partial name match (server-side) | false |
| `--rate-limit` |  | `0` | Filter VXCs by rate limit in Mbps | false |
| `--status` |  |  | Filter VXCs by status (comma-separated, e.g. LIVE,CONFIGURED) | false |

## Subcommands
* [docs](megaport-cli_vxc_list_docs.md)

