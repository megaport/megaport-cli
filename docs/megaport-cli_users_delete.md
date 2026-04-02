# delete

Delete a user

## Description

Delete a user from your Megaport company.

This command deletes a user by their employee ID. Only users with pending invitations can be deleted. Users who have already logged in must be deactivated instead.

### Important Notes
  - Only users with pending invitations can be deleted
  - To remove access for active users, use the 'deactivate' command instead

### Example Usage

```sh
  megaport-cli users delete 12345
  megaport-cli users delete 12345 --force
```

## Usage

```sh
megaport-cli users delete [flags]
```


## Parent Command

* [megaport-cli users](megaport-cli_users.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Skip the confirmation prompt | false |

## Subcommands
* [docs](megaport-cli_users_delete_docs.md)

