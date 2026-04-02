# status

Check the provisioning status of an IX

## Description

Check the provisioning status of an IX through the Megaport API.

This command retrieves only the essential status information for an Internet Exchange (IX) without all the details. It's useful for monitoring ongoing provisioning.

### Important Notes
  - This is a lightweight command that only shows the IX's status without retrieving all details.

### Example Usage

```sh
  megaport-cli ix status ix-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## Usage

```sh
megaport-cli ix status [flags]
```


## Parent Command

* [megaport-cli ix](megaport-cli_ix.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_ix_status_docs.md)

