# update-prefix-filter-list

Update a prefix filter list on an MCR

## Description

Update a prefix filter list on an MCR.

This command allows you to update the details of an existing prefix filter list on an MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each field you want to update.

2. Flag Mode:
   Provide fields as flags:
   --description, --address-family, --entries

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- `description`: The new description of the prefix filter list.
- `address_family`: The new address family (IPv4/IPv6).
- `entries`: JSON array of prefix filter entries. Each entry has:
- `action`: "permit" or "deny"
- `prefix`: CIDR notation (e.g., "192.168.0.0/16")
- `ge` (optional): Greater than or equal to value
- `le` (optional): Less than or equal to value

Example usage:

### Interactive mode
```
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive

```

### Flag mode
```
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --description "Updated prefix list" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

```

### JSON mode
```
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{"description":"Updated prefix list","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json

```



## Usage

```
megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --address-family |  |  | New address family (IPv4 or IPv6) | false |
| --description |  |  | New description of the prefix filter list | false |
| --entries |  |  | JSON array of prefix filter entries | false |
| --interactive | -i | false | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing prefix filter list configuration | false |
| --json-file |  |  | Path to JSON file containing prefix filter list configuration | false |



