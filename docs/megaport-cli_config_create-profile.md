# create-profile

Create a new credential profile

## Description

Create a new profile with Megaport API credentials and environment settings.

Profiles store your Megaport API access and secret keys along with environment settings for secure reuse. The profile name is case-sensitive and must be unique.

Credentials are stored in ~/.megaport/config.json with secure file permissions.

If --access-key or --secret-key are not provided, you will be prompted. On an interactive terminal input is masked; on piped/non-TTY stdin it is read without masking.

### Important Notes
  - API credentials are stored with 0600 permissions (readable only by the current user)
  - Passing --access-key or --secret-key on the command line exposes credentials in shell history and process listings. Omit them to be prompted securely, or use env vars MEGAPORT_ACCESS_KEY / MEGAPORT_SECRET_KEY instead. Note: the secure prompt masks input only on an interactive terminal; piped input is read without masking.

### Example Usage

```sh
  megaport-cli config create-profile production --environment production
  megaport-cli config create-profile staging --environment staging --description "Staging credentials"
```

## Usage

```sh
megaport-cli config create-profile [flags]
```


## Parent Command

* [megaport-cli config](megaport-cli_config.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--access-key` |  |  | Megaport API access key (omit to be prompted; masked on TTY only) | false |
| `--description` |  |  | Optional description for this profile | false |
| `--environment` |  | `production` | Target API environment: 'production', 'staging', or 'development' | false |
| `--secret-key` |  |  | Megaport API secret key (omit to be prompted; masked on TTY only) | false |

## Subcommands
* [docs](megaport-cli_config_create-profile_docs.md)

