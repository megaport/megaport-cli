# topology

Show resource relationship tree

## Description

Show a tree view of Megaport resources and their VXC connections.

This command fetches all Ports, MCRs, and MVEs and renders each with its associated Virtual Cross Connects (VXCs) as a tree. The B-End destination of each VXC is shown to illustrate connectivity.

Default output is a human-readable ASCII tree. Use --output json for structured output.

### Important Notes
  - Each VXC is shown once, under its A-End parent resource
  - CSV and XML output formats are not supported for hierarchical topology data

### Example Usage

```sh
  megaport-cli topology
  megaport-cli topology --output json
  megaport-cli topology --type mcr
  megaport-cli topology --include-inactive
```

## Usage

```sh
megaport-cli topology [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--include-inactive` |  | `false` | Include deprovisioned resources in the tree | false |
| `--type` |  |  | Filter by resource type: port, mcr, or mve | false |

## Subcommands
* [docs](megaport-cli_topology_docs.md)

