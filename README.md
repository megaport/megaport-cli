# Megaport CLI

## Overview

The Megaport CLI provides a command-line interface for managing Megaport resources and services. It allows users to interact with the Megaport API directly from their terminal.

Before using this CLI, please ensure you read and accept Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/) and [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy/).

For API details, consult the [Megaport API Documentation](https://dev.megaport.com/).

## Installation

```sh
# Install using Go
go install github.com/megaport/megaport-cli@latest

# Verify installation
megaport --version
```

## Shell Completion

The CLI supports shell completion for bash, zsh, fish, and PowerShell:

```sh
# Bash (Linux)
megaport completion bash > /etc/bash_completion.d/megaport

# Bash (macOS with Homebrew)
megaport completion bash > $(brew --prefix)/etc/bash_completion.d/megaport

# Zsh
megaport completion zsh > "${fpath[1]}/_megaport"

# Fish
megaport completion fish > ~/.config/fish/completions/megaport.fish

# PowerShell
megaport completion powershell > megaport.ps1
```

## Configuration

Configure your CLI credentials using either environment variables or the configure command:

```sh
# Using configure command
megaport configure --access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY

# Using environment variables
export MEGAPORT_ACCESS_KEY=<your-access-key>
export MEGAPORT_SECRET_KEY=<your-secret-key>
export MEGAPORT_ENVIRONMENT=<environment>  # production, staging, or development
```

## Available Commands

### Resource Management
- `locations`: List and search Megaport locations
- `ports`: Manage Megaport ports
- `mcr`: Manage Megaport Cloud Routers
- `mve`: Manage Megaport Virtual Edge instances
- `vxc`: Manage Virtual Cross Connects
- `partners`: List and search partner ports
- `servicekeys`: Manage service keys

### Output Formats
All commands support multiple output formats:
- `--output table` (default)
- `--output json`
- `--output csv`

### Examples

```sh
# List all locations
megaport locations list

# Get port details
megaport ports get PORT_UID --output json

# List partner ports with filtering
megaport partners list \
  --product-name "AWS Direct Connect" \
  --connect-type "AWSHC" \
  --output table

# Create a service key
megaport servicekeys create \
  --product-uid PRODUCT_UID \
  --description "My Service Key" \
  --max-speed 1000
```

## Environment Support

The CLI supports different Megaport environments:
- Production (default)
- Staging
- Development

Specify the environment using the `--env` flag or `MEGAPORT_ENVIRONMENT` variable.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to the project.

All contributions are subject to the [Megaport Contributor License Agreement](CLA.md) and [Mozilla Public License 2.0](LICENSE).

## Support

- [Open Issues](https://github.com/megaport/megaport-cli/issues)
- [API Documentation](https://dev.megaport.com/)
- [Megaport Website](https://www.megaport.com)