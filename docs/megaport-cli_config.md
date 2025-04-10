# config

Manage configuration settings

## Description

Manage configuration settings for Megaport CLI.

The config command allows you to manage persistent configuration settings for the CLI, including authentication profiles with environment settings. Profiles store your API credentials and environment settings for streamlined operations across multiple Megaport environments.

Configuration is stored locally in ~/.megaport/config.json and persists across CLI sessions.

Configuration Precedence:
1. Command-line flags (highest precedence)
2. Environment variables (MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, etc.)
3. Active profile in config file
4. Default settings in config file (lowest precedence)

### Important Notes
  - Configuration contains sensitive credentials - ensure ~/.megaport directory has appropriate permissions
  - Environment variables (MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY) take precedence over stored profiles

### Example Usage

```sh
  megaport-cli config create-profile production --access-key xxx --secret-key xxx --environment production
  megaport-cli config use-profile production
```

## Usage

```sh
megaport-cli config [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [clear-defaults](megaport-cli_config_clear-defaults.md)
* [create-profile](megaport-cli_config_create-profile.md)
* [delete-profile](megaport-cli_config_delete-profile.md)
* [docs](megaport-cli_config_docs.md)
* [export](megaport-cli_config_export.md)
* [get-default](megaport-cli_config_get-default.md)
* [import](megaport-cli_config_import.md)
* [list-profiles](megaport-cli_config_list-profiles.md)
* [remove-default](megaport-cli_config_remove-default.md)
* [set-default](megaport-cli_config_set-default.md)
* [update-profile](megaport-cli_config_update-profile.md)
* [use-profile](megaport-cli_config_use-profile.md)
* [view](megaport-cli_config_view.md)

