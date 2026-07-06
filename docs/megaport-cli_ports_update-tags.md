# update-tags

Update resource tags on a specific port

## Description

Update resource tags associated with a specific port. Tags can be provided via interactive prompts, JSON string, or JSON file.

## Usage

```sh
megaport-cli ports update-tags [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Skip the confirmation prompt | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |

