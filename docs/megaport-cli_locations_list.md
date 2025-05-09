# list

List all locations

## Description

List all locations available in the Megaport API.

This command fetches and displays a list of all available locations with details such as location ID, name, country, and metro. You can also filter the locations based on specific criteria.

### Optional Fields
  - `country`: Filter locations by country
  - `metro`: Filter locations by metro area
  - `name`: Filter locations by name

### Example Usage

```sh
  megaport-cli locations list
  megaport-cli locations list --metro "San Francisco"
  megaport-cli locations list --country "US"
  megaport-cli locations list --name "Equinix SY1"
```

## Usage

```sh
megaport-cli locations list [flags]
```


## Parent Command

* [megaport-cli locations](megaport-cli_locations.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--country` |  |  | Filter locations by country | false |
| `--metro` |  |  | Filter locations by metro area | false |
| `--name` |  |  | Filter locations by name | false |

## Subcommands
* [docs](megaport-cli_locations_list_docs.md)

