# restore

Restore a deleted port

## Description

Restore a previously deleted port in the Megaport API.

This command allows you to restore a port that has been marked for deletion but not yet fully decommissioned. The port will be reinstated with its original configuration.

### Important Notes
  - You can only restore ports that are in a "DECOMMISSIONING" state
  - Once a port is fully decommissioned, it cannot be restored
  - The restoration process is immediate but may take a few minutes to complete
  - All port attributes will be restored to their pre-deletion state
  - You will resume being billed for the port according to your original terms

### Example Usage

```sh
  megaport-cli ports restore 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
```

## Usage

```sh
megaport-cli ports restore [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_ports_restore_docs.md)

