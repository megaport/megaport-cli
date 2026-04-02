# deactivate

Deactivate a user

## Description

Deactivate a user in your Megaport company.

This command deactivates a user by setting their active status to false. The user will no longer be able to log in or perform actions.

### Important Notes
  - Deactivated users cannot log in or perform any actions
  - The user's email will be modified to prevent reuse

### Example Usage

```sh
  megaport-cli users deactivate 12345
  megaport-cli users deactivate 12345 --force
```

## Usage

```sh
megaport-cli users deactivate [flags]
```


## Parent Command

* [megaport-cli users](megaport-cli_users.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Skip the confirmation prompt | false |

## Subcommands
* [docs](megaport-cli_users_deactivate_docs.md)

