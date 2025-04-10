# import

Import configuration

## Description

Import configuration from a file.

Import allows you to load profiles and default settings from a JSON file. The import file must follow the structure of an export file, with valid credentials in place of any [REDACTED] values.

IMPORTANT: Importing merges with existing configuration. It will:
- Add new profiles that don't exist
- Update existing profiles with the same name
- Add or update default settings
- Set the active profile if specified in the import file

Version compatibility: Import supports config file versions up to the current version.

### Required Fields
  - `file`: File to import from

### Important Notes
  - Credentials marked as [REDACTED] in export files must be replaced with actual values before import

### Example Usage

```sh
  megaport-cli config import --file myconfig.json
```

## Usage

```sh
megaport-cli config import [flags]
```


## Parent Command

* [megaport-cli config](megaport-cli_config.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--file` |  |  | File to import from | true |

## Subcommands
* [docs](megaport-cli_config_import_docs.md)

