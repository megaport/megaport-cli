# validate-lag

Validate a LAG port order without purchasing

## Description

Validates a LAG port configuration against the Megaport API without creating the resource.

Use this for dry-run validation before purchasing, or in CI pipelines to check configurations.

### Required Fields
  - `lag-count`: The number of LAG members (between 1 and 8)
  - `location-id`: The ID of the location where the port will be provisioned
  - `marketplace-visibility`: Whether the port should be visible in the marketplace (true or false)
  - `name`: The name of the port (1-64 characters)
  - `port-speed`: The speed of each LAG member port (10000 or 100000 Mbps)
  - `term`: The term of the port (1, 12, or 24 months)

### Optional Fields
  - `resource-tags`: Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"})
  - `resource-tags-file`: Path to JSON file containing resource tags

### Important Notes
  - This command only validates the configuration — no resources are created and no charges are incurred

### Example Usage

```sh
  megaport-cli ports validate-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
  megaport-cli ports validate-lag --json-file ./lag-config.json
```

## Usage

```sh
megaport-cli ports validate-lag [flags]
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
| `--lag-count` |  | `0` | The number of LAG members (between 1 and 8) | true |
| `--location-id` |  | `0` | The ID of the location where the port will be provisioned | true |
| `--marketplace-visibility` |  | `false` | Whether the port should be visible in the marketplace (true or false) | true |
| `--name` |  |  | The name of the port (1-64 characters) | true |
| `--port-speed` |  | `0` | The speed of each LAG member port (10000 or 100000 Mbps) | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--resource-tags` |  |  | Resource tags as a JSON string (e.g. {"key1":"value1","key2":"value2"}) | false |
| `--resource-tags-file` |  |  | Path to JSON file containing resource tags | false |
| `--term` |  | `0` | The term of the port (1, 12, or 24 months) | true |

## Subcommands
* [docs](megaport-cli_ports_validate-lag_docs.md)

