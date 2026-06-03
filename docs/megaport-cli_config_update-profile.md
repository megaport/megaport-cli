# update-profile

Update an existing profile

## Description

Update an existing profile with new credentials or settings.

To avoid recording the secret value in shell history, pass an empty string (e.g. --secret-key "") and you will be prompted instead of providing the value on the command line. On an interactive terminal input is masked; on piped/non-TTY stdin it is read without masking. Alternatively, use env vars MEGAPORT_ACCESS_KEY / MEGAPORT_SECRET_KEY which always take precedence over stored profiles.

### Important Notes
  - Keep your Megaport API credentials secure; they provide full account access
  - Passing --access-key or --secret-key on the command line exposes credentials in shell history and process listings. Pass an empty value to be prompted instead (masked on a TTY; echoed on piped/non-TTY stdin).

### Example Usage

```sh
  megaport-cli config update-profile myprofile --environment staging
  megaport-cli config update-profile myprofile --secret-key ""
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
| `--access-key` |  |  | New Megaport API access key (pass empty string to be prompted; masked on TTY only) | false |
| `--description` |  |  | Profile description (use empty string to clear) | false |
| `--environment` |  |  | Target API environment: 'production', 'staging', or 'development' | false |
| `--secret-key` |  |  | New Megaport API secret key (pass empty string to be prompted; masked on TTY only) | false |

## Subcommands
* [docs](megaport-cli_config_update-profile_docs.md)

