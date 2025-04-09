# buy

Purchase a new Megaport Virtual Edge (MVE) device

## Description

Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.

This command allows you to purchase an MVE by providing the necessary details.

### Required Fields
  - `location-id`: The ID of the location where the MVE will be provisioned
  - `name`: The name of the MVE
  - `term`: The term of the MVE (1, 12, 24, or 36 months)
  - `vendor-config`: JSON string with vendor-specific configuration (for flag mode)
  - `vnics`: JSON array of network interfaces (for flag mode)

### Optional Fields
  - `cost-centre`: Cost centre for billing
  - `diversity-zone`: The diversity zone for the MVE
  - `promo-code`: Promotional code for discounts

### Important Notes
  - For production deployments, you may want to use a JSON file to manage complex configurations
  - To list available images and their IDs, use: megaport-cli mve list-images
  - To list available sizes, use: megaport-cli mve list-sizes
  - Location IDs can be retrieved with: megaport-cli locations list

### Example Usage

```sh
  megaport-cli mve buy --interactive
  megaport-cli mve buy --name "My MVE" --term 12 --location-id 123 --vendor-config '{"vendor":"cisco","imageId":123,"productSize":"MEDIUM"}' --vnics '[{"description":"Data Plane","vlan":100}]'
  megaport-cli mve buy --json '{"name":"My MVE","term":12,"locationId":123,"vendorConfig":{"vendor":"cisco","imageId":123,"productSize":"MEDIUM"},"vnics":[{"description":"Data Plane","vlan":100}]}'
  megaport-cli mve buy --json-file ./mve-config.json
```
### JSON Format Example
```json
{
  "name": "My MVE Display Name",
  "term": 12,
  "locationId": 123,
  "diversityZone": "blue",
  "promoCode": "PROMO2023",
  "costCentre": "Marketing Dept",
  "vendorConfig": {
    "vendor": "cisco",
    "imageId": 123,
    "productSize": "MEDIUM",
    "mveLabel": "custom-label",
    "manageLocally": true,
    "adminSshPublicKey": "ssh-rsa AAAA...",
    "sshPublicKey": "ssh-rsa AAAA...",
    "cloudInit": "#cloud-config\npackages:\n - nginx\n"
  },
  "vnics": [
    {"description": "Data Plane", "vlan": 100},
    {"description": "Management", "vlan": 200}
  ]
}

```

## Usage

```sh
megaport-cli mve buy [flags]
```


## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)
## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | The diversity zone for the MVE | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing configuration | false |
| `--json-file` |  |  | Path to JSON file containing configuration | false |
| `--location-id` |  | `0` | The ID of the location where the MVE will be provisioned | true |
| `--name` |  |  | The name of the MVE | true |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | The term of the MVE (1, 12, 24, or 36 months) | true |
| `--vendor-config` |  |  | JSON string with vendor-specific configuration (for flag mode) | true |
| `--vnics` |  |  | JSON array of network interfaces (for flag mode) | true |

## Subcommands
* [docs](megaport-cli_mve_buy_docs.md)

