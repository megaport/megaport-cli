# create-profile

Create a new credential profile

## Description

Create a new profile with Megaport API credentials and environment settings.

Profiles store your Megaport API access and secret keys along with environment settings for secure reuse. The profile name is case-sensitive and must be unique.

Credentials are stored in ~/.megaport/config.json with secure file permissions.

### Required Fields
  - `access-key`: Megaport API access key from the Megaport Portal
  - `secret-key`: Megaport API secret key from the Megaport Portal

### Important Notes
  - API credentials are stored with 0600 permissions (readable only by the current user)

### Example Usage

```sh
  megaport-cli config create-profile production --access-key xxx --secret-key xxx --environment production --description "Production credentials"
  megaport-cli config create-profile staging --access-key yyy --secret-key yyy --environment staging
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
| `--access-key` |  |  | Megaport API access key from the Megaport Portal | true |
| `--description` |  |  | Optional description for this profile | false |
| `--environment` |  | `production` | Target API environment: 'production', 'staging', or 'development' | false |
| `--secret-key` |  |  | Megaport API secret key from the Megaport Portal | true |

## Subcommands
* [docs](megaport-cli_config_create-profile_docs.md)

