# status

Check the provisioning status of a port

## Description

Check the provisioning status of a port through the Megaport API.

This command retrieves only the essential status information for a port without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the port's status without retrieving all details.

### Example Usage

```sh
  megaport-cli ports status port-abc123
```

## Usage

```sh
megaport-cli ports status [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_ports_status_docs.md)

