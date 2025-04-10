# list

List all MVEs with optional filters

## Description

List all MVEs available in the Megaport API.

This command fetches and displays a list of MVEs with details such as MVE ID, name, location, vendor, and status. By default, only active MVEs are shown.

### Example Usage

```sh
  megaport-cli mve list
  megaport-cli mve list --location-id 123
  megaport-cli mve list --vendor "Cisco"
  megaport-cli mve list --name "Edge Router"
  megaport-cli mve list --include-inactive
  megaport-cli mve list --location-id 123 --vendor "Cisco" --name "Edge"
```

## Usage

```sh
megaport-cli mve list [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive MVEs in the list | false |
| `--location-id` |  | `0` | Filter MVEs by location ID | false |
| `--name` |  |  | Filter MVEs by name | false |
| `--vendor` |  |  | Filter MVEs by vendor | false |

## Subcommands
* [docs](megaport-cli_mve_list_docs.md)

