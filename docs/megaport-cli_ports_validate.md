# validate

Validate a port order without purchasing

## Description

Validates a port configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `location-id`: The ID of the location where the port will be provisioned
  - `marketplace-visibility`: Whether the port should be visible in the marketplace (true or false)
  - `name`: The name of the port (1-64 characters)
  - `port-speed`: The speed of the port (1000, 10000, or 100000 Mbps)
  - `term`: The term of the port (1, 12, 24, or 36 months)

### Optional Fields
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli ports validate --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
  megaport-cli ports validate --json-file ./port-config.json
```

## Usage

```sh
megaport-cli ports validate [flags]
```


## Parent Command

* [megaport-cli ports](megaport-cli_ports.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | Diversity zone for the port | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The ID of the location where the port will be provisioned | true |
| `--marketplace-visibility` |  | `false` | Whether the port should be visible in the marketplace (true or false) | true |
| `--name` |  |  | The name of the port (1-64 characters) | true |
| `--port-speed` |  | `0` | The speed of the port (1000, 10000, or 100000 Mbps) | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--term` |  | `0` | The term of the port (1, 12, 24, or 36 months) | true |

## Subcommands
* [docs](megaport-cli_ports_validate_docs.md)

