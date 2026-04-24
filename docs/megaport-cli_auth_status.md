# status

Display current authentication status and identity

## Description

Verify your credentials and display the current account context.

This command authenticates with the Megaport API using your active profile or environment variables, then retrieves your company user details. It shows the company, environment, active profile, and API endpoint.

Note: the displayed user is inferred from the company user list (preferring the primary admin). For companies with multiple admins, it may not reflect the exact user who owns the API credentials.

Use this to confirm which account and environment you are operating against before making infrastructure changes.

### Example Usage

```sh
  megaport-cli auth status
  megaport-cli auth status --output json
  megaport-cli auth status --output json --query '[0].email'
```

## Usage

```sh
megaport-cli auth status [flags]
```


## Parent Command

* [megaport-cli auth](megaport-cli_auth.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|

## Subcommands
* [docs](megaport-cli_auth_status_docs.md)

