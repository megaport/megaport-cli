# list

List all MVEs with optional filters

## Description

List all MVEs available in the Megaport API.

This command fetches and displays a list of MVEs with details such as MVE ID, name, location, vendor, and status. By default, only active MVEs are shown. You can also filter by resource tags.

### Optional Fields
  - `tag`: Filter by resource tag (format: key=value or key; repeatable, AND logic)

### Example Usage

```sh
  megaport-cli mve list
  megaport-cli mve list --location-id 123
  megaport-cli mve list --vendor "Cisco"
  megaport-cli mve list --name "Edge Router"
  megaport-cli mve list --include-inactive
  megaport-cli mve list --location-id 123 --vendor "Cisco" --name "Edge"
  megaport-cli mve list --tag env=prod
  megaport-cli mve list --tag env=prod --tag team=network
```

## Usage

```sh
megaport-cli mve list [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include inactive MVEs in the list | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--location-id` |  | `0` | Filter MVEs by location ID | false |
| `--name` |  |  | Filter MVEs by name | false |
| `--tag` |  | `[]` | Filter by resource tag (format: key=value or key; repeatable, AND logic) | false |
| `--vendor` |  |  | Filter MVEs by vendor | false |

