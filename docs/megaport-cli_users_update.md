# update

Update an existing user

## Description

Update an existing user's details.

This command allows you to update specific properties of an existing user. Only provided fields will be changed.

### Optional Fields
  - `active`: Set user active status
  - `email`: New email address
  - `first-name`: New first name
  - `last-name`: New last name
  - `notification-enabled`: Enable/disable notifications
  - `phone`: New phone number
  - `position`: New position/role

### Important Notes
  - Users with pending invitations cannot be updated

### Example Usage

```sh
  megaport-cli users update 12345 --interactive
  megaport-cli users update 12345 --first-name "Jane" --last-name "Smith"
  megaport-cli users update 12345 --json '{"firstName":"Jane"}'
```

## Usage

```sh
megaport-cli users update [flags]
```


## Parent Command

* [megaport-cli users](megaport-cli_users.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--active` |  | `false` | Set user active status | false |
| `--email` |  |  | New email address | false |
| `--first-name` |  |  | New first name | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--last-name` |  |  | New last name | false |
| `--notification-enabled` |  | `false` | Enable/disable notifications | false |
| `--phone` |  |  | New phone number | false |
| `--position` |  |  | New position/role | false |

## Subcommands
* [docs](megaport-cli_users_update_docs.md)

