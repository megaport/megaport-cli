# add-ipsec-addon

Add an IPSec add-on to an existing MCR

## Description

Add an IPSec add-on to an existing MCR.

This command provisions an IPSec add-on on the specified MCR. IPSec add-ons enable encrypted tunnel termination on the MCR.

### Optional Fields
  - `tunnel-count`: Number of IPSec tunnels (10, 20, or 30); omit or set to 0 to use the API default of 10

### Important Notes
  - Valid tunnel counts are 10, 20, or 30. Omit --tunnel-count or set to 0 to use the API default (10).
  - You must provide one of: --tunnel-count, --interactive, --json, or --json-file

### Example Usage

```sh
  megaport-cli mcr add-ipsec-addon [mcrUID] --tunnel-count 10
  megaport-cli mcr add-ipsec-addon [mcrUID] --tunnel-count 20
  megaport-cli mcr add-ipsec-addon [mcrUID] --json '{"tunnelCount":10}'
  megaport-cli mcr add-ipsec-addon [mcrUID] --interactive
```
### JSON Format Example
```json
{
  "tunnelCount": 10
}

```

## Usage

```sh
megaport-cli mcr add-ipsec-addon [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--tunnel-count` |  | `0` | IPSec tunnel count (10, 20, or 30; omit or use 0 to let the API apply its default of 10) | false |

## Subcommands
* [docs](megaport-cli_mcr_add-ipsec-addon_docs.md)

