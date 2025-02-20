# Megaport CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/megaport/megaport-cli.svg)](https://pkg.go.dev/github.com/megaport/megaport-cli)

## Overview

This is the Megaport CLI. It allows users to interact with the Megaport API from the command line.

Before using this CLI, please ensure you read Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/).

The [Megaport API Documentation](https://dev.megaport.com/) is also available online.

## Getting started 

```sh
# Install the Megaport CLI
go install github.com/megaport/megaport-cli@latest

# Configure the CLI with your credentials
mp1 configure --access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY

# List all available locations
mp1 locations list
```

## Contributing

Contributions via pull request are welcome. Familiarize yourself with these guidelines to increase the likelihood of your pull request being accepted.

All contributions are subject to the [Megaport Contributor Licence Agreement](CLA.md).
The CLA clarifies the terms of the [Mozilla Public Licence 2.0](LICENSE) used to Open Source this repository and ensures that contributors are explicitly informed of the conditions. Megaport requires all contributors to accept these terms to ensure that the Megaport CLI remains available and licensed for the community.

The main themes of the [Megaport Contributor Licence Agreement](CLA.md) cover the following conditions: 
- Clarifying the Terms of the [Mozilla Public Licence 2.0](LICENSE), used to Open Source this project.
- As a contributor, you have permission to agree to the License terms.
- As a contributor, you are not obligated to provide support or warranty for your contributions.
- Copyright is assigned to Megaport to use as Megaport determines, including within commercial products.
- Grant of Patent Licence to Megaport for any contributions containing patented or future patented works.

The [Megaport Contributor Licence Agreement](CLA.md) is the authoritative document over these conditions and any other communications unless explicitly stated otherwise.

When you open a Pull Request, all authors of the contributions are required to comment on the Pull Request confirming acceptance of the CLA terms. Pull Requests cannot be merged until this is complete.

The [Megaport Contributor Licence Agreement](CLA.md) applies to contributions. 
All users are free to use the `megaport-cli` project under the [MPL-2.0 Open Source Licence](LICENSE).

Megaport users are also bound by the [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy).	

### Getting Started

Prior to working on new code, review the [Open Issues](../issues). Check whether your issue has already been raised, and consider working on an issue with votes or clear demand.

If you don't see an open issue for your need, open one and let others know what you are working on. Avoid lengthy or complex changes that rewrite the repository or introduce breaking changes. Straightforward pull requests based on discussion or ideas and Megaport feedback are the most likely to be accepted. 

Megaport is under no obligation to accept any pull requests or to accept them in full. You are free to fork and modify the code for your own use as long as it is published under the MPL-2.0 License.