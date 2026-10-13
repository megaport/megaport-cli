# list

List locations

## Description

List locations available in the Megaport API.

This command fetches and displays locations with details such as location ID, name, country, and metro. It shows only Active locations unless you pass --status. You can also filter the locations based on specific criteria.

### Optional Fields
  - `country`: Filter locations by country
  - `market-code`: Filter locations by market code
  - `mcr-available`: Filter locations that support MCR
  - `metro`: Filter locations by metro area
  - `name`: Filter locations by name
  - `status`: Filter locations by status: Active, Deployment, or Restricted (repeatable, default Active)

### Example Usage

```sh
  megaport-cli locations list
  megaport-cli locations list --metro "San Francisco"
  megaport-cli locations list --country "US"
  megaport-cli locations list --name "Equinix SY1"
  megaport-cli locations list --status Restricted --status Deployment
```

## Usage

```sh
megaport-cli locations list [flags]
```


## Parent Command

* [megaport-cli locations](megaport-cli_locations.md)

## Aliases

* ls
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--country` |  |  | Filter locations by country | false |
| `--limit` |  | `0` | Maximum number of results to display (0 = unlimited) | false |
| `--market-code` |  |  | Filter locations by market code | false |
| `--mcr-available` |  | `false` | Filter locations that support MCR | false |
| `--metro` |  |  | Filter locations by metro area | false |
| `--name` |  |  | Filter locations by name | false |
| `--status` |  | `[]` | Filter locations by status: Active, Deployment, or Restricted (repeatable, default Active) | false |

