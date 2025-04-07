# create-prefix-filter-list

Create a prefix filter list on an MCR

## Description

Create a prefix filter list on an MCR.

This command allows you to create a new prefix filter list on an MCR. Prefix filter lists are used to control which routes are accepted or advertised by the MCR.

Required fields:
  address-family: The address family (IPv4 or IPv6)
  description: The description of the prefix filter list (1-255 characters)
  entries: JSON array of prefix filter entries

Important notes:
  - The address_family must be either "IPv4" or "IPv6"
  - The entries must be a valid JSON array of prefix filter entries
  - The ge and le values are optional but must be within the range of the prefix length

Example usage:

```
  create-prefix-filter-list [mcrUID] --interactive
  create-prefix-filter-list [mcrUID] --description "My prefix list" --address-family "IPv4" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'
  create-prefix-filter-list [mcrUID] --json '{"description":"My prefix list","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json
```
JSON format example:
```
{
  "description": "My prefix list",
  "addressFamily": "IPv4",
  "entries": [
    {
      "action": "permit",
      "prefix": "10.0.0.0/8",
      "ge": 24,
      "le": 32
    },
    {
      "action": "deny",
      "prefix": "0.0.0.0/0"
    }
  ]
}
```


## Usage

```
megaport-cli mcr create-prefix-filter-list [mcrUID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--address-family` |  |  | Address family (IPv4 or IPv6) | false |
| `--description` |  |  | Description of the prefix filter list | false |
| `--entries` |  |  | JSON array of prefix filter entries | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing prefix filter list configuration | false |
| `--json-file` |  |  | Path to JSON file containing prefix filter list configuration | false |



