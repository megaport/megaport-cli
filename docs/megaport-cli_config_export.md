# export

Export configuration

## Description

Export configuration to a file (excluding sensitive information).

The export function writes your configuration to a JSON file with sensitive information like access keys and secret keys REDACTED for security purposes. This means you CANNOT directly import an export file to restore credentials.

Exports are useful for:
- Backing up profile settings and defaults (without credentials)
- Sharing configuration templates with teammates
- Transferring settings between environments

To use an exported file on another system, you must manually edit the file to replace [REDACTED] values with actual credentials before importing.

### Example Usage

```sh
  megaport-cli config export --file myconfig.json
```

## Usage

```sh
megaport-cli config export [flags]
```


## Parent Command

* [megaport-cli config](megaport-cli_config.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--file` |  |  | File to export to | false |

## Subcommands
* [docs](megaport-cli_config_export_docs.md)

