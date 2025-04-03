# update

Update an existing MCR

## Description

Update an existing Megaport Cloud Router (MCR).

This command allows you to update the details of an existing MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each field you want to update.

2. Flag Mode:
   Provide fields as flags:
   --name, --cost-centre, --marketplace-visibility, --term

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- `name`: The new name of the MCR.
- `cost_centre`: The new cost center for the MCR.
- `marketplace_visibility`: The new marketplace visibility (true/false).
- `term`: The new contract term in months (1, 12, 24, or 36).

Example usage:

  # Interactive mode
```
  megaport-cli mcr update [mcrUID] --interactive
```

  # Flag mode
```
  megaport-cli mcr update [mcrUID] --name "Updated MCR" --marketplace-visibility true
```

  # JSON mode
```
  megaport-cli mcr update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true}'
  megaport-cli mcr update [mcrUID] --json-file ./update-mcr-config.json
```



## Usage

```
megaport-cli mcr update [mcrUID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --cost-centre |  |  | Cost centre for billing | false |
| --interactive | -i | false | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing MCR configuration | false |
| --json-file |  |  | Path to JSON file containing MCR configuration | false |
| --marketplace-visibility |  | false | Whether the MCR is visible in marketplace | false |
| --name |  |  | New MCR name | false |
| --term |  | 0 | New contract term in months (1, 12, 24, or 36) | false |



