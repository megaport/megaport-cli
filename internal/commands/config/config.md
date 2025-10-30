# Megaport CLI Configuration System

This document provides a comprehensive overview of the Megaport CLI configuration system.

> **⚠️ WASM Note**: Configuration profiles are **NOT available** in the WASM/browser version of Megaport CLI.
> The WASM version uses **session-based authentication** managed through the browser UI.
> Please use the login form in the web interface for authentication instead of config commands.
>
> This documentation applies only to the standard CLI (non-WASM) version.

## Configuration File

The configuration file is stored at `~/.megaport/config.json` by default, or at the path specified by the `MEGAPORT_CONFIG_DIR` environment variable. The file has 0600 permissions (readable/writable only by the owner) to protect sensitive credential information.

### Structure

```json
{
  "version": 1,
  "active_profile": "myprofile",
  "profiles": {
    "myprofile": {
      "access_key": "key123",
      "secret_key": "secret456",
      "environment": "production",
      "description": "My profile"
    },
    "staging-profile": {
      "access_key": "key789",
      "secret_key": "secret101112",
      "environment": "staging",
      "description": "Staging environment"
    }
  },
  "defaults": {
    "output": "table",
    "no-color": false
  }
}
```

### Key Components

- **version**: Config file format version (currently 2)
- **active_profile**: Name of the currently selected profile
- **profiles**: Map of profile names to profile configurations
  - **access_key**: Megaport API access key
  - **secret_key**: Megaport API secret key
  - **environment**: API environment to use (`production`, `staging`, or `development`)
  - **description**: Optional user-provided description
- **defaults**: Map of default settings for CLI operation

## Configuration Precedence

Settings are applied in the following order (highest to lowest precedence):

1. **Command-line flags**: Flags provided directly to a command always have highest priority
2. **Environment variables**: `MEGAPORT_ACCESS_KEY`, `MEGAPORT_SECRET_KEY`, etc.
3. **Active profile**: Settings from the active profile in the config file
4. **Default settings**: Values in the `defaults` section of the config file

## Profile Management

### Creating Profiles

Profiles store API credentials and environment settings for convenient reuse. Create them with:

```
megaport-cli config create-profile myprofile --access-key xxx --secret-key xxx --environment production
```

### Switching Profiles

Change the active profile with:

```
megaport-cli config use-profile myprofile
```

### Updating Profiles

Update an existing profile with:

```
megaport-cli config update-profile myprofile --access-key xxx --environment staging
```

Only specified fields will be updated, others remain unchanged.

### Deleting Profiles

Delete a profile with:

```
megaport-cli config delete-profile myprofile
```

**Note**: You cannot delete the currently active profile.

### Listing Profiles

List all available profiles with:

```
megaport-cli config list-profiles
```

## Default Settings

Default settings apply when no profile setting or command-line flag is provided.

### Setting Defaults

```
megaport-cli config set-default output json
```

### Getting Defaults

```
megaport-cli config get-default output
```

### Removing Defaults

```
megaport-cli config remove-default output
```

### Clearing All Defaults

```
megaport-cli config clear-defaults
```

## Import and Export

### Exporting Configuration

Export your configuration to a file with:

```
megaport-cli config export --file myconfig.json
```

**Important**: For security reasons, sensitive information like access keys and secret keys are **REDACTED** in exports.

### Importing Configuration

Import configuration from a file with:

```
megaport-cli config import --file myconfig.json
```

**Note**: For security reasons, imported files must have actual values (not `[REDACTED]`) for credential fields. This typically means you'll need to manually edit exported files before importing them elsewhere.

Import behavior:

- Adds new profiles that don't exist
- Updates existing profiles with the same name
- Adds or updates default settings
- Sets the active profile if specified in the import file

## Security Considerations

- The config file uses 0600 permissions (readable/writable only by the file owner)
- API credentials are sensitive and provide account access - protect them accordingly
- Exported configurations have redacted credentials by design
- Consider using environment variables for CI/CD pipelines instead of stored profiles

## Edge Cases

- **Empty profile names**: Not allowed
- **Whitespace-only profile names**: Not allowed
- **Special characters in profile names**: Supported, including Unicode characters
- **Case sensitivity**: Profile names are case-sensitive (`myprofile` vs `MyProfile`)
- **Very long paths**: The configuration system supports deeply nested paths for MEGAPORT_CONFIG_DIR
- **Symlinks**: The configuration system properly handles symbolic links for the config directory

## Troubleshooting

- **Corrupted config file**: If your config file becomes corrupted, the system will create a new default config
- **Permission issues**: Ensure you have read/write permissions on ~/.megaport
- **Environment variable conflicts**: Remember that environment variables take precedence over profile settings
