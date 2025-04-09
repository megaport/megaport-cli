# create-prefix-filter-list

Create a prefix filter list on an MCR

## Description

Create a prefix filter list on an MCR.

This command allows you to create a new prefix filter list on an MCR. Prefix filter lists are used to control which routes are accepted or advertised by the MCR.

### Important Notes
  - The address_family must be either "IPv4" or "IPv6"
  - The entries must be a valid JSON array of prefix filter entries
  - The ge and le values are optional but must be within the range of the prefix length
  - Required flags (description, address-family, entries) can be skipped when using --interactive, --json, or --json-file

### Example Usage

```sh
  create-prefix-filter-list [mcrUID] --interactive
  create-prefix-filter-list [mcrUID] --description "My prefix list" --address-family "IPv4" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'
  create-prefix-filter-list [mcrUID] --json '{"description":"My prefix list","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json
```
### JSON Format Example
```json
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

```sh
megaport-cli mcr create-prefix-filter-list [flags]
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
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |


