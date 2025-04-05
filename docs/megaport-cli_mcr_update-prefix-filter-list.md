# update-prefix-filter-list

Update a prefix filter list on an MCR

## Description

Update a prefix filter list on an MCR.

This command allows you to update the details of an existing prefix filter list on an MCR. You can use this command to modify the description, address family, or entries in the list.

Optional fields:
entries: JSON array of prefix filter entries
description: The new description of the prefix filter list (1-255 characters)
address-family: The new address family (IPv4 or IPv6)

Example usage:

update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive
update-prefix-filter-list [mcrUID] [prefixFilterListID] --description "Updated prefix list" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'
update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{"description":"Updated prefix list","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json

JSON format example:
{
"description": "Updated prefix list",
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



## Usage

```
megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--address-family` |  |  | New address family (IPv4 or IPv6) | false |
| `--description` |  |  | New description of the prefix filter list | false |
| `--entries` |  |  | JSON array of prefix filter entries | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing prefix filter list configuration | false |
| `--json-file` |  |  | Path to JSON file containing prefix filter list configuration | false |



