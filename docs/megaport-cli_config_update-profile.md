# update-profile

Update an existing profile

## Description

Update an existing profile with new credentials or settings.

### Important Notes
  - Keep your Megaport API credentials secure; they provide full account access

### Example Usage

```sh
  megaport-cli config update-profile myprofile --access-key xxx --secret-key xxx
  megaport-cli config update-profile myprofile --environment staging
```

## Usage

```sh
megaport-cli config update-profile [flags]
```


## Parent Command

* [megaport-cli config](megaport-cli_config.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--access-key` |  |  | Megaport API access key | false |
| `--description` |  |  | Profile description (use empty string to clear) | false |
| `--environment` |  |  | Target API environment: 'production', 'staging', or 'development' | false |
| `--secret-key` |  |  | Megaport API secret key | false |

## Subcommands
* [docs](megaport-cli_config_update-profile_docs.md)

