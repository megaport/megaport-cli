# list

List all locations

## Description

List all locations available in the Megaport API.

This command fetches and displays a list of all available locations with details such as location ID, name, country, and metro. You can also filter the locations based on specific criteria.

Optional fields:
metro: Filter locations by metro area
country: Filter locations by country
name: Filter locations by name

Example usage:

list
list --metro "San Francisco"
list --country "US"
list --name "Equinix SY1"



## Usage

```
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



