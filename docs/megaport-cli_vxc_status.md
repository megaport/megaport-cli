# status

Check the provisioning status of a VXC

## Description

Check the provisioning status of a VXC through the Megaport API.

This command retrieves only the essential status information for a Virtual Cross Connect (VXC) without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the VXC's status without retrieving all details.

### Example Usage

```sh
  megaport-cli vxc status vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  megaport-cli vxc status vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --watch
  megaport-cli vxc status vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --watch --interval 10s
```

## Usage

```sh
megaport-cli vxc status [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)

## Aliases

* st
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--interval` |  | `5s` | Polling interval for --watch mode (e.g. 5s, 1m) | false |
| `--watch` | `-w` | `false` | Continuously poll and display resource status (Ctrl+C to stop) | false |

## Subcommands
* [docs](megaport-cli_vxc_status_docs.md)

