# list

List all company users

## Description

List all users in your Megaport company.

This command fetches and displays a list of users with details such as employee ID, name, email, position, and active status.

### Optional Fields
  - `active-only`: Show only active users
  - `inactive-only`: Show only inactive users
  - `position`: Filter users by position/role

### Example Usage

```sh
  megaport-cli users list
  megaport-cli users list --active-only
  megaport-cli users list --position "Technical Admin"
```

## Usage

```sh
megaport-cli users list [flags]
```


## Parent Command

* [megaport-cli users](megaport-cli_users.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--active-only` |  | `false` | Show only active users | false |
| `--inactive-only` |  | `false` | Show only inactive users | false |
| `--position` |  |  | Filter users by position/role | false |

## Subcommands
* [docs](megaport-cli_users_list_docs.md)

