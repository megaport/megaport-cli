# status

Check the provisioning status of an MVE

## Description

Check the provisioning status of an MVE through the Megaport API.

This command retrieves only the essential status information for a Megaport Virtual Edge (MVE) without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the MVE's status without retrieving all details.

### Example Usage

```sh
  megaport-cli mve status mve-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## Usage

```sh
megaport-cli mve status [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_mve_status_docs.md)

