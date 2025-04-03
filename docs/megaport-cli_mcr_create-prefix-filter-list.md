# create-prefix-filter-list

Create a prefix filter list on an MCR

## Description

Create a prefix filter list on an MCR.

This command allows you to create a new prefix filter list on an MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required field.

2. Flag Mode:
   Provide all required fields as flags:
   --description, --address-family, --entries

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
- `description`: The description of the prefix filter list.
- `address_family`: The address family (IPv4/IPv6).
- `entries`: JSON array of prefix filter entries. Each entry has:
- `action`: "permit" or "deny"
- `prefix`: CIDR notation (e.g., "192.168.0.0/16")
- `ge` (optional): Greater than or equal to value
- `le` (optional): Less than or equal to value

Example usage:

### Interactive mode
```
  megaport-cli mcr create-prefix-filter-list [mcrUID] --interactive

```

### Flag mode
```
  megaport-cli mcr create-prefix-filter-list [mcrUID] --description "My prefix list" --address-family "IPv4" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

```

### JSON mode
```
  megaport-cli mcr create-prefix-filter-list [mcrUID] --json '{"description":"My prefix list","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport-cli mcr create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json

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
| --address-family |  |  | Address family (IPv4 or IPv6) | false |
| --description |  |  | Description of the prefix filter list | false |
| --entries |  |  | JSON array of prefix filter entries | false |
| --interactive | -i | false | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing prefix filter list configuration | false |
| --json-file |  |  | Path to JSON file containing prefix filter list configuration | false |



