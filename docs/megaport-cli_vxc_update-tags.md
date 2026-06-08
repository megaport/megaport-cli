# update-tags

Update resource tags on a specific VXC

## Description

Update resource tags associated with a specific VXC. Tags can be provided via interactive prompts, JSON string, or JSON file.

## Usage

```sh
megaport-cli vxc update-tags [flags]
```


## Parent Command

* [megaport-cli vxc](megaport-cli_vxc.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Skip the confirmation prompt | false |
| `--generate-skeleton` |  | `false` | Print a JSON skeleton template for --json or --json-file input and exit | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |

## Subcommands
* [docs](megaport-cli_vxc_update-tags_docs.md)

