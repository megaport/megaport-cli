# Megaport CLI

> [!CAUTION]
> The Megaport CLI tool is currently an unsupported alpha, we're excited for feedback but please know that functionality and features may change drastically, and there may be bugs.

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

## Environment Support

The CLI supports different Megaport environments:
- Production (default)
- Staging
- Development

Specify the environment using the `--env` flag or `MEGAPORT_ENVIRONMENT` variable.

## Configuration

Configure your CLI credentials using environment variables.

```sh
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

#### Locations
```sh
# List all locations
megaport locations list

# List locations filtered by metro area
megaport locations list --metro "San Francisco"

# Get details for a specific location
megaport locations get LOCATION_ID --output json
```

#### Ports
```sh
# List all ports
megaport ports list

# List ports filtered by location ID and port speed
megaport ports list --location-id 1 --port-speed 10000

# Get details for a specific port
megaport ports get PORT_UID --output json

# Buy a new port
megaport ports buy --interactive
megaport ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
megaport ports buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
megaport ports buy --json-file ./port-config.json

# Buy a LAG port
megaport ports buy-lag --interactive
megaport ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
megaport ports buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
megaport ports buy-lag --json-file ./lag-port-config.json

# Update a port
megaport ports update PORT_UID --interactive
megaport ports update PORT_UID --name "Updated Port" --marketplace-visibility true
megaport ports update PORT_UID --json '{"name":"Updated Port","marketplaceVisibility":true}'
megaport ports update PORT_UID --json-file ./update-port-config.json

# Delete a port
megaport ports delete PORT_UID --now

# Restore a deleted port
megaport ports restore PORT_UID

# Lock a port
megaport ports lock PORT_UID

# Unlock a port
megaport ports unlock PORT_UID

# Check VLAN availability on a port
megaport ports check-vlan PORT_UID VLAN_ID
```

#### MCR (Megaport Cloud Routers)
```sh
# Get details for a specific MCR
megaport mcr get MCR_UID --output json

# Buy a new MCR
megaport mcr buy

megaport mcr update

# Delete an MCR
megaport mcr delete MCR_UID --now

# Restore a deleted MCR
megaport mcr restore MCR_UID

# Create a prefix filter list on an MCR
megaport mcr create-prefix-filter-list MCR_UID

# List all prefix filter lists for a specific MCR
megaport mcr list-prefix-filter-lists MCR_UID

# Get details for a specific prefix filter list on an MCR
megaport mcr get-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID

# Update a prefix filter list on an MCR
megaport mcr update-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID

# Delete a prefix filter list on an MCR
megaport mcr delete-prefix-filter-list MCR_UID PREFIX_FILTER_LIST_ID
```

#### MVE (Megaport Virtual Edge)
```sh
# Get details for a specific MVE
megaport mve get MVE_UID --output json

# Buy a new MVE
megaport mve buy

# Update an existing MVE
megaport mve update MVE_UID

# Delete an MVE
megaport mve delete --now

# List all available MVE images
megaport mve list-images

# List MVE images filtered by vendor
megaport mve list-images --vendor "Cisco"

# List MVE images filtered by product code
megaport mve list-images --product-code "FORTINET456"

# List MVE images filtered by ID
megaport mve list-images --id 1

# List MVE images filtered by version
megaport mve list-images --version "2.0"

# List MVE images filtered by release image
megaport mve list-images --release-image

# List all available MVE sizes
megaport mve list-sizes
```

#### VXC (Virtual Cross Connects)
```sh
# Get details for a specific VXC
megaport vxc get VXC_UID --output json

# Buy a new VXC
megaport vxc buy

# Delete a VXC
megaport vxc delete VXC_UID
```

#### Partners
```sh
# List all partner ports
megaport partners list

# List partner ports filtered by product name and company name
megaport partners list --product-name "AWS Direct Connect" --company-name "Acme Corp"

# Get details for a specific partner port
megaport partners get PARTNER_UID --output json
```

#### Service Keys
```sh
# List all service keys
megaport servicekeys list

# Get details for a specific service key
megaport servicekeys get SERVICE_KEY_UID --output json

# Create a new service key
megaport servicekeys create --product-uid PRODUCT_UID --description "My Service Key" --max-speed 1000

# Update an existing service key
megaport servicekeys update SERVICE_KEY_UID --description "Updated Description"
```

## Contributing

Contributions via pull request are welcome. Familiarize yourself with these guidelines to increase the likelihood of your pull request being accepted.

All contributions are subject to the [Megaport Contributor Licence Agreement](CLA.md).

The CLA clarifies the terms of the [Mozilla Public Licence 2.0](LICENSE) used to Open Source this repository and ensures that contributors are explicitly informed of the conditions. Megaport requires all contributors to accept these terms to ensure that the Megaport Terraform Provider remains available and licensed for the community.

When you open a Pull Request, all authors of the contributions are required to comment on the Pull Request confirming acceptance of the CLA terms. Pull Requests cannot be merged until this is complete.

Megaport users are also bound by the [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy).

## Support

- [Open Issues](https://github.com/megaport/megaport-cli/issues)
- [API Documentation](https://dev.megaport.com/)
- [Megaport Website](https://www.megaport.com)
