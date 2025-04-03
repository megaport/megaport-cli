# update

Update a port's details

## Description

Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --marketplace-visibility

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
- `name`: The new name of the port.
- `marketplace_visibility`: Whether the port should be visible in the marketplace (true or false).

Optional fields:
- `cost_centre`: The cost center for the port.
- `term`: The new term of the port (1, 12, or 24 months).

Example usage:

### Interactive mode
```
  megaport-cli ports update [portUID] --interactive

```

### Flag mode
```
  megaport-cli ports update [portUID] --name "Updated Port" --marketplace-visibility true

```

### JSON mode
```
  megaport-cli ports update [portUID] --json '{"name":"Updated Port","marketplaceVisibility":true}'
  megaport-cli ports update [portUID] --json-file ./update-port-config.json

```



## Usage

```
megaport-cli ports update [portUID] [flags]
```



## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing port configuration | false |
| `--json-file` |  |  | Path to JSON file containing port configuration | false |
| `--marketplace-visibility` |  | `false` | Whether the port is visible in marketplace | false |
| `--name` |  |  | New port name | false |
| `--term` |  | `0` | New contract term in months (1, 12, 24, or 36) | false |



