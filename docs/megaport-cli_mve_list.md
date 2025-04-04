# list

List all MVEs

## Description

List all Megaport Virtual Edge (MVE) devices associated with your Megaport account.

This command retrieves all MVEs from the Megaport API and displays them in the specified format.
By default, inactive MVEs are excluded. Use the --inactive flag to include them.

You can filter the results using these options:
--name         Filter by name (case-insensitive substring match)
--location-id  Filter by location ID
--vendor       Filter by vendor name

Example usage:

### List all active MVEs
```
megaport-cli mve list

```
### List all MVEs including inactive ones
```
megaport-cli mve list --inactive

```
### Filter MVEs by name
```
megaport-cli mve list --name "production"

```
### Filter MVEs by location ID
```
megaport-cli mve list --location-id 67

```
### Filter MVEs by vendor
```
megaport-cli mve list --vendor "cisco"

```
### Combine multiple filters
```
megaport-cli mve list --vendor "cisco" --location-id 67 --inactive

```
### List all MVEs in JSON format
```
megaport-cli mve list --output json

```


## Usage

```
megaport-cli mve list [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--inactive` |  | `false` | Include inactive MVEs in the list | false |
| `--location-id` |  | `0` | Filter MVEs by location ID | false |
| `--name` |  |  | Filter MVEs by name (case-insensitive substring match) | false |
| `--vendor` |  |  | Filter MVEs by vendor | false |



