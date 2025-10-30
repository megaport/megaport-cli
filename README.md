# Megaport CLI

> [!CAUTION]
> The Megaport CLI tool is currently an unsupported alpha, we're excited for feedback but please know that functionality and features may change drastically, and there may be bugs.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Getting Started](#getting-started)
- [Environment Support](#environment-support)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Shell Completion](#shell-completion)
- [Architecture](#architecture)
- [Available Commands](#available-commands)
- [Troubleshooting](#troubleshooting)
- [Additional Documentation](#additional-documentation)
- [Contributing](#contributing)
- [Support](#support)

## Overview

The Megaport CLI provides a command-line interface for managing Megaport resources and services. It allows users to interact with the Megaport API directly from their terminal.

### 🌐 WebAssembly Browser Version Available!

Try the Megaport CLI directly in your browser - no installation required! The WASM version runs entirely in your browser and can be deployed with Docker.

👉 **[See WASM_README.md](WASM_README.md) for deployment instructions**

**Quick Deploy:**

```bash
./deploy.sh
# Then open http://localhost:8080
```

---

Before using this CLI, please ensure you read and accept Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/) and [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy/).

For API details, consult the [Megaport API Documentation](https://dev.megaport.com/).

## Installation

```sh
# Install using Go
go install github.com/megaport/megaport-cli@latest

# Verify installation
megaport-cli version
```

## Getting Started

To quickly begin using the Megaport CLI:

```sh
# Set up your environment variables
export MEGAPORT_ACCESS_KEY=your_access_key
export MEGAPORT_SECRET_KEY=your_secret_key

# Test your configuration
megaport-cli locations list

# Get help for any command
megaport-cli vxc --help
```

## Documentation

Generate comprehensive documentation for all CLI commands:

```sh
# Generate documentation in the docs directory
megaport-cli generate-docs ./docs
```

## Shell Completion

The CLI supports shell completion for bash, zsh, fish, and PowerShell:

```sh
# Bash (Linux)
megaport-cli completion bash > /etc/bash_completion.d/megaport

# Bash (macOS with Homebrew)
megaport-cli completion bash > $(brew --prefix)/etc/bash_completion.d/megaport

# Zsh
megaport-cli completion zsh > "${fpath[1]}/_megaport"

# Fish
megaport-cli completion fish > ~/.config/fish/completions/megaport.fish

# PowerShell
megaport-cli completion powershell > megaport.ps1
```

## Environment Support

The CLI supports different Megaport environments:

- Production (default)
- Staging
- Development

Specify the environment using the `--env` flag or `MEGAPORT_ENVIRONMENT` variable.

## Configuration

The Megaport CLI uses a profile-based configuration system that securely stores API credentials and default settings.

### Key Features

- **Profiles**: Store multiple sets of credentials for different environments
- **Secure Storage**: Credentials stored with 0600 permissions in `~/.megaport/config.json`
- **Default Settings**: Configure preferred output format and other settings

### Basic Usage

```bash
# Create a profile
megaport-cli config create-profile myprofile --access-key xxx --secret-key xxx

# Switch active profile
megaport-cli config use-profile myprofile

# View current configuration
megaport-cli config view
```

### Configuration Precedence

Settings are applied in this order (highest to lowest precedence):

1. Command-line flags (e.g., `--access-key`)
2. Environment variables (`MEGAPORT_ACCESS_KEY`, etc.)
3. Active profile in config file
4. Default settings in config file

### Environment Variables

For CI/CD pipelines or temporary usage:

```sh
export MEGAPORT_ACCESS_KEY=<your-access-key>
export MEGAPORT_SECRET_KEY=<your-secret-key>
export MEGAPORT_ENVIRONMENT=<environment>  # production, staging, or development
```

For complete documentation on configuration options, profile management, import/export functionality, and troubleshooting, see the [Configuration Guide](internal/commands/config/config.md).

> **Note**: Configuration profiles are only available in the standard CLI version. The WASM/browser version uses session-based authentication via the web UI login form.

## Architecture

The Megaport CLI is built using a Command Builder pattern that ensures consistent behavior across all commands. This modular approach makes the CLI easy to extend and maintain while providing a consistent user experience.

Key architectural components:

- **Command Builder**: Declarative command definitions for consistent behavior
- **Flag Sets**: Reusable sets of flags organized by resource type
- **Output Formatters**: Consistent data presentation across all commands

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

#### Locations

```sh
# List all locations
megaport-cli locations list

# List locations filtered by metro area
megaport-cli locations list --metro "San Francisco"

# Get details for a specific location
megaport-cli locations get LOCATION_ID --output json
```

#### Ports

```sh
# List all ports
megaport-cli ports list

# List ports filtered by location ID and port speed
megaport-cli ports list --location-id 1 --port-speed 1000

# Get details for a specific port
megaport-cli ports get PORT_UID --output json

# Buy a new port
megaport-cli ports buy --interactive
megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
megaport-cli ports buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
megaport-cli ports buy --json-file ./port-config.json

# Buy a LAG port
megaport-cli ports buy-lag --interactive
megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
megaport-cli ports buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
megaport-cli ports buy-lag --json-file ./lag-port-config.json

# Update a port
megaport-cli ports update PORT_UID --interactive
megaport-cli ports update PORT_UID --name "Updated Port" --marketplace-visibility true
megaport-cli ports update PORT_UID --json '{"name":"Updated Port","marketplaceVisibility":true}'
megaport-cli ports update PORT_UID --json-file ./update-port-config.json

# Delete a port
megaport-cli ports delete PORT_UID --now

# Restore a deleted port
megaport-cli ports restore PORT_UID

# Lock a port
megaport-cli ports lock PORT_UID

# Unlock a port
megaport-cli ports unlock PORT_UID

# Check VLAN availability on a port
megaport-cli ports check-vlan PORT_UID VLAN_ID
```

#### MCR (Megaport Cloud Routers)

```sh
# List all MCRs
megaport-cli mcr list

# Get details for a specific MCR
megaport-cli mcr get MCR_UID --output json

# Buy a new MCR
megaport-cli mcr buy --interactive
megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 10000 --location-id 123
megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 10000 --location-id 123 --mcr-asn 65000 --diversity-zone "blue"
megaport-cli mcr buy --json '{"name":"My MCR","term":12,"portSpeed":10000,"locationId":123}'
megaport-cli mcr buy --json-file ./mcr-config.json

# Update an MCR
megaport-cli mcr update MCR_UID --interactive
megaport-cli mcr update MCR_UID --name "Updated MCR" --cost-centre "IT-123"
megaport-cli mcr update MCR_UID --json '{"name":"Updated MCR","costCentre":"IT-123"}'
megaport-cli mcr update MCR_UID --json-file ./update-mcr-config.json

# Delete an MCR
megaport-cli mcr delete MCR_UID --now

# Restore a deleted MCR
megaport-cli mcr restore MCR_UID

# Create a prefix filter list on an MCR
megaport-cli mcr create-prefix-filter-list MCR_UID --interactive
megaport-cli mcr create-prefix-filter-list MCR_UID --description "Block List" --address-family "IPv4" --entries '[{"action":"DENY","prefix":"10.0.0.0/8","ge":16,"le":24}]'
megaport-cli mcr create-prefix-filter-list MCR_UID --json '{"description":"Block List","addressFamily":"IPv4","entries":[{"action":"DENY","prefix":"10.0.0.0/8","ge":16,"le":24}]}'
megaport-cli mcr create-prefix-filter-list MCR_UID --json-file ./prefix-filter-config.json

# List all prefix filter lists for a specific MCR
megaport-cli mcr list-prefix-filter-lists MCR_UID

# Get details for a specific prefix filter list on an MCR
megaport-cli mcr get-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID --output json

# Update a prefix filter list on an MCR
megaport-cli mcr update-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID --interactive
megaport-cli mcr update-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID --description "Updated Block List" --address-family "IPv4" --entries '[{"action":"DENY","prefix":"10.0.0.0/8"},{"action":"PERMIT","prefix":"192.168.0.0/16"}]'
megaport-cli mcr update-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID --json '{"description":"Updated Block List","addressFamily":"IPv4","entries":[{"action":"DENY","prefix":"10.0.0.0/8"},{"action":"PERMIT","prefix":"192.168.0.0/16"}]}'
megaport-cli mcr update-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID --json-file ./update-prefix-filter-config.json

# Delete a prefix filter list on an MCR
megaport-cli mcr delete-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID
```

#### MVE (Megaport Virtual Edge)

```sh
# Get details for a specific MVE
megaport-cli mve get MVE_UID --output json

# Buy a new MVE
megaport-cli mve buy --interactive

# Buy a new MVE - Cisco example
megaport-cli mve buy --name "My Cisco MVE" --term 12 --location-id 67 --vendor-config '{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"cisco-mve","manageLocally":true,"adminSshPublicKey":"ssh-rsa AAAA...","sshPublicKey":"ssh-rsa AAAA...","cloudInit":"#cloud-config\npackages:\n - nginx\n","fmcIpAddress":"10.0.0.1","fmcRegistrationKey":"key123","fmcNatId":"natid123"}' --vnics '[{"description":"Data Plane","vlan":100}]'
megaport-cli mve buy --json '{"name":"My Cisco MVE","term":12,"locationId":67,"vendorConfig":{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"cisco-mve","manageLocally":true,"adminSshPublicKey":"ssh-rsa AAAA...","sshPublicKey":"ssh-rsa AAAA...","cloudInit":"#cloud-config\npackages:\n - nginx\n","fmcIpAddress":"10.0.0.1","fmcRegistrationKey":"key123","fmcNatId":"natid123"},"vnics":[{"description":"Data Plane","vlan":100}]}'

# Buy a new MVE - Aruba example
megaport-cli mve buy --name "My Aruba MVE" --term 1 --location-id 67 --vendor-config '{"vendor":"aruba","imageId":23,"productSize":"MEDIUM","accountName":"Aruba Test Account","accountKey":"12345678","systemTag":"Preconfiguration-aruba-test-1"}' --vnics '[{"description":"Data Plane"},{"description":"Control Plane"},{"description":"Management Plane"}]'
megaport-cli mve buy --json '{"name":"My Aruba MVE","term":1,"locationId":67,"vendorConfig":{"vendor":"aruba","imageId":23,"productSize":"MEDIUM","accountName":"Aruba Test Account","accountKey":"12345678","systemTag":"Preconfiguration-aruba-test-1"},"vnics":[{"description":"Data Plane"},{"description":"Control Plane"},{"description":"Management Plane"}]}'

# Buy a new MVE - Versa example
megaport-cli mve buy --name "My Versa MVE" --term 1 --location-id 67 --vendor-config '{"vendor":"versa","imageId":20,"productSize":"MEDIUM","directorAddress":"director1.versa.com","controllerAddress":"controller1.versa.com","localAuth":"SDWAN-Branch@Versa.com","remoteAuth":"Controller-1-staging@Versa.com","serialNumber":"Megaport-Hub1"}' --vnics '[{"description":"Data Plane"}]'
megaport-cli mve buy --json '{"name":"My Versa MVE","term":1,"locationId":67,"vendorConfig":{"vendor":"versa","imageId":20,"productSize":"MEDIUM","directorAddress":"director1.versa.com","controllerAddress":"controller1.versa.com","localAuth":"SDWAN-Branch@Versa.com","remoteAuth":"Controller-1-staging@Versa.com","serialNumber":"Megaport-Hub1"},"vnics":[{"description":"Data Plane"}]}'

# Update an existing MVE
megaport-cli mve update MVE_UID --interactive
megaport-cli mve update MVE_UID --name "Updated MVE Name" --cost-centre "New Cost Centre" --contract-term 24
megaport-cli mve update MVE_UID --json '{"name": "New MVE Name", "costCentre": "New Cost Centre", "contractTermMonths": 24}'

# Delete an MVE
megaport-cli mve delete MVE_UID --now

# List all available MVE images
megaport-cli mve list-images

# List MVE images filtered by vendor
megaport-cli mve list-images --vendor "Cisco"

# List MVE images filtered by product code
megaport-cli mve list-images --product-code "FORTINET456"

# List MVE images filtered by ID
megaport-cli mve list-images --id 1

# List MVE images filtered by version
megaport-cli mve list-images --version "2.0"

# List MVE images filtered by release image
megaport-cli mve list-images --release-image

# List all available MVE sizes
megaport-cli mve list-sizes
```

#### VXC (Virtual Cross Connects)

```sh
# Get details for a specific VXC
megaport-cli vxc get VXC_UID --output json

# Buy a new VXC - Interactive mode.
megaport-cli vxc buy --interactive

# Flag mode - Basic VXC between two ports
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-vlan 200

# Flag mode - VXC to AWS Direct Connect
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My AWS VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AWS","ownerAccount":"123456789012","asn":65000,"amazonAsn":64512}'

# Flag mode - VXC to Azure ExpressRoute
megaport-cli vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --name "My Azure VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AZURE","serviceKey":"s-abcd1234"}'

# Interactive mode
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive

# Flag mode - Basic updates
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --name "New VXC Name" \
  --rate-limit 2000 \
  --cost-centre "New Cost Centre"

# Flag mode - Update VLANs
megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --a-end-vlan 200 \
  --b-end-vlan 300

# Delete a VXC
megaport-cli vxc delete VXC_UID
```

#### Partners

```sh
# List all partner ports
megaport-cli partners list

# List partner ports filtered by product name and company name
megaport-cli partners list --product-name "AWS Direct Connect" --company-name "AWS"

# Interactive search for partner ports with prompts
megaport-cli partners find
```

#### Service Keys

```sh
# List all service keys
megaport-cli servicekeys list

# Get details for a specific service key
megaport-cli servicekeys get SERVICE_KEY_UID --output json

# Create a new service key
megaport-cli servicekeys create --product-uid PRODUCT_UID --description "My Service Key" --max-speed 1000

# Update an existing service key
megaport-cli servicekeys update SERVICE_KEY_UID --description "Updated Description"
```

## Troubleshooting

Common issues and their solutions:

- **Authentication errors**: Ensure your access key and secret key are correctly set and have the necessary permissions
- **Rate limiting**: The Megaport API has rate limits; if you encounter 429 errors, add delays between requests
- **Output formatting issues**: Use the `--no-color` flag if terminal colors are causing display problems
- **Missing resources**: Resources may take time to provision; use appropriate wait times in automation scripts

### Workflow Example: Set up a Cloud Connection

This example shows setting up a complete cloud connection workflow:

```sh
# 1. Find available locations
megaport-cli locations list --metro "Sydney" --output json # Find the location ID from this output.

# 2. Create a port at your chosen location
megaport-cli ports buy --name "SYD-Port-1" --location-id 15 --port-speed 1000 --term 12

# 3. Find your cloud provider's connection point
megaport-cli partners list --company-name "AWS" --connect-type "AWS" --location-id 15

# 4. Create a VXC to connect your port to AWS
megaport-cli vxc buy --a-end-uid "port-xxxxxxxx" \
  --name "AWS-Connection" \
  --rate-limit 500 \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AWS","ownerAccount":"123456789012", "type":"private"}'
```

## Additional Documentation

The CLI includes comprehensive generated documentation in the `docs` folder:

- **[Index of all commands](docs/index.md)** - Complete listing of all available commands
- **Command References:**
  - Locations - Find and explore Megaport locations
  - Ports - Manage physical ports
  - VXC - Virtual cross connects between endpoints
  - MCR - Megaport Cloud Router management
  - MVE - Megaport Virtual Edge devices
  - Partners - Cloud service provider connections
  - Service Keys - Manage service keys for connections

Each command page includes:

- Detailed descriptions
- All available flags and options
- Usage examples
- Links to related subcommands

You can regenerate this documentation at any time with:

```sh
megaport-cli generate-docs ./docs
```

## Contributing

Contributions via pull request are welcome. Familiarize yourself with these guidelines to increase the likelihood of your pull request being accepted.

All contributions are subject to the Megaport Contributor Licence Agreement.

The CLA clarifies the terms of the Mozilla Public Licence 2.0 used to Open Source this repository and ensures that contributors are explicitly informed of the conditions. Megaport requires all contributors to accept these terms to ensure that the Megaport Terraform Provider remains available and licensed for the community.

When you open a Pull Request, all authors of the contributions are required to comment on the Pull Request confirming acceptance of the CLA terms. Pull Requests cannot be merged until this is complete.

Megaport users are also bound by the [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy/).

## Support

- [Open Issues](https://github.com/megaport/megaport-cli/issues)
- [API Documentation](https://dev.megaport.com/)
- [Megaport Website](https://www.megaport.com)
