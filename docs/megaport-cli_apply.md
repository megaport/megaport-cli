# apply

Provision multiple resources from a config file

## Description

Provision multiple Megaport resources (ports, MCRs, MVEs, VXCs) from a declarative YAML or JSON config file.

Resources are provisioned sequentially in dependency order: ports and MCRs first, then MVEs, then VXCs. VXC endpoints can reference previously provisioned resources using {{.type.name}} template syntax.

### Required Fields
  - `file`: Path to config file (YAML or JSON)

### Example Usage

```sh
  megaport apply -f infrastructure.yaml
  megaport apply -f infrastructure.yaml --dry-run
  megaport apply -f infrastructure.yaml --yes
  megaport apply -f infrastructure.json --output json
```

## Usage

```sh
megaport-cli apply [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--dry-run` |  | `false` | Validate all orders without provisioning | false |
| `--file` | `-f` |  | Path to config file (YAML or JSON) | true |
| `--yes` | `-y` | `false` | Skip confirmation prompt | false |

## Subcommands
* [docs](megaport-cli_apply_docs.md)

