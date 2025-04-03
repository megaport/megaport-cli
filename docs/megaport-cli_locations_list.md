# list

List all locations

## Description

List all locations available in the Megaport API.

This command fetches and displays a list of all available locations with details such as
location ID, name, country, and region. You can also filter the locations based on specific criteria.

Available filters:
  - metro: Filter locations by metro area.
  - country: Filter locations by country.
  - name: Filter locations by name.

Example usage:

  megaport-cli locations list
  megaport-cli locations list --metro "San Francisco"
  megaport-cli locations list --country "US"
  megaport-cli locations list --name "Equinix"



## Usage

```
megaport-cli locations list [flags]
```



## Parent Command

* [megaport-cli locations](locations.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --country |  |  | Filter locations by country | false |
| --metro |  |  | Filter locations by metro area | false |
| --name |  |  | Filter locations by name | false |



