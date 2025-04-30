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
```

## Usage

```sh
megaport-cli vxc status [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_vxc_status_docs.md)

