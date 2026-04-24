# update-ipsec-addon

Update or disable an IPSec add-on on an MCR

## Description

Update or disable an existing IPSec add-on on an MCR.

This command updates the tunnel count on an existing IPSec add-on. Set tunnel-count to 0 to disable the IPSec add-on.

### Required Fields
  - `tunnel-count`: New tunnel count (10, 20, or 30); set to 0 to disable IPSec

### Important Notes
  - Valid tunnel counts are 10, 20, or 30. Set to 0 to disable the IPSec add-on.
  - --tunnel-count can be skipped when using --interactive, --json, or --json-file

### Example Usage

```sh
  megaport-cli mcr update-ipsec-addon [mcrUID] [addOnUID] --tunnel-count 20
  megaport-cli mcr update-ipsec-addon [mcrUID] [addOnUID] --tunnel-count 0
  megaport-cli mcr update-ipsec-addon [mcrUID] [addOnUID] --json '{"tunnelCount":30}'
  megaport-cli mcr update-ipsec-addon [mcrUID] [addOnUID] --interactive
```
### JSON Format Example
```json
{
  "tunnelCount": 30
}

```

## Usage

```sh
megaport-cli mcr update-ipsec-addon [flags]
```


## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--tunnel-count` |  | `0` | New tunnel count (10, 20, or 30); set to 0 to disable IPSec | true |

## Subcommands
* [docs](megaport-cli_mcr_update-ipsec-addon_docs.md)

