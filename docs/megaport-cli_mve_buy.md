# buy

Purchase a new Megaport Virtual Edge (MVE) device

## Description

Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.

This command allows you to purchase an MVE by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
 The command will prompt you for each required and optional field.

2. Flag Mode:
 Provide all required fields as flags:
 --name, --term, --location-id, --vendor-config, --vnics

3. JSON Mode:
 Provide a JSON string or file with all required fields:
 --json <json-string> or --json-file <path>

Required fields:
- name: The name of the MVE.
- term: The term of the MVE (1, 12, 24, or 36 months).
- location_id: The ID of the location where the MVE will be provisioned.
- vendor-config: JSON string with vendor-specific configuration (for flag mode)
- vnics: JSON array of network interfaces (for flag mode)

Vendor-specific fields:
- 6WIND:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ssh_public_key (required)
- Aruba:
- image_id (required)
- product_size (required)
- mve_label (optional)
- account_name (required)
- account_key (required)
- system_tag (optional)
- Aviatrix:
- image_id (required)
- product_size (required)
- mve_label (optional)
- cloud_init (required)
- Cisco:
- image_id (required)
- product_size (required)
- mve_label (required)
- manage_locally (required, true/false)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- cloud_init (required)
- fmc_ip_address (required)
- fmc_registration_key (required)
- fmc_nat_id (required)
- Fortinet:
- image_id (required)
- product_size (required)
- mve_label (optional)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- ha_license (required for HA deployments)
- license_data (required for non-HA deployments)
- PaloAlto:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ssh_public_key (required)
- admin_password_hash (required)
- license_data (required)
- Prisma:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ion_key (required)
- secret_key (required)
- Versa:
- image_id (required)
- product_size (required)
- mve_label (optional)
- director_address (required)
- controller_address (required)
- local_auth (required)
- remote_auth (required)
- serial_number (required)
- VMware:
- image_id (required)
- product_size (required)
- mve_label (optional)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- vco_address (required)
- vco_activation_code (required)
- Meraki:
- image_id (required)
- product_size (required)
- mve_label (optional)
- token (required)

Example usage:

### Interactive mode
```
megaport-cli mve buy --interactive
```

### Flag mode - Cisco example
```
megaport-cli mve buy --name "My Cisco MVE" --term 12 --location-id 123 \
  --vendor-config '{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"cisco-mve",
                   "manageLocally":true,"adminSshPublicKey":"ssh-rsa AAAA...","sshPublicKey":"ssh-rsa AAAA...",
                   "cloudInit":"#cloud-config\npackages:\n - nginx\n","fmcIpAddress":"10.0.0.1",
                   "fmcRegistrationKey":"key123","fmcNatId":"natid123"}' \
  --vnics '[{"description":"Data Plane","vlan":100}]'
```

### Flag mode - Aruba example
```
megaport-cli mve buy --name "Megaport MVE Example" --term 1 --location-id 123 \
  --vendor-config '{"vendor":"aruba","imageId":23,"productSize":"MEDIUM",
                   "accountName":"Aruba Test Account","accountKey":"12345678",
                   "systemTag":"Preconfiguration-aruba-test-1"}' \
  --vnics '[{"description":"Data Plane"},{"description":"Control Plane"},{"description":"Management Plane"}]'
```

### Flag mode - Versa example
```
megaport-cli mve buy --name "Megaport Versa MVE Example" --term 1 --location-id 123 \
  --vendor-config '{"vendor":"versa","imageId":20,"productSize":"MEDIUM",
                   "directorAddress":"director1.versa.com","controllerAddress":"controller1.versa.com",
                   "localAuth":"SDWAN-Branch@Versa.com","remoteAuth":"Controller-1-staging@Versa.com",
                   "serialNumber":"Megaport-Hub1"}' \
  --vnics '[{"description":"Data Plane"}]'
```


### JSON mode - Cisco example
```
megaport-cli mve buy --json '{
"name": "My Cisco MVE",
"term": 12,
"locationId": 67,
"vendorConfig": {
  "vendor": "cisco",
  "imageId": 1,
  "productSize": "large",
  "mveLabel": "cisco-mve",
  "manageLocally": true,
  "adminSshPublicKey": "ssh-rsa AAAA...",
  "sshPublicKey": "ssh-rsa AAAA...",
  "cloudInit": "#cloud-config\npackages:\n - nginx\n",
  "fmcIpAddress": "10.0.0.1",
  "fmcRegistrationKey": "key123",
  "fmcNatId": "natid123"
},
"vnics": [
  {"description": "Data Plane", "vlan": 100}
]
}'
```

### JSON mode - Aruba example
```
megaport-cli mve buy --json '{
"name": "Megaport MVE Example",
"term": 1,
"locationId": 67,
"vendorConfig": {
  "vendor": "aruba",
  "imageId": 23,
  "productSize": "MEDIUM",
  "accountName": "Aruba Test Account",
  "accountKey": "12345678",
  "systemTag": "Preconfiguration-aruba-test-1"
},
"vnics": [
  {"description": "Data Plane"},
  {"description": "Control Plane"},
  {"description": "Management Plane"}
]
}'
```

### JSON mode - Versa example
```
megaport-cli mve buy --json '{
"name": "Megaport Versa MVE Example",
"term": 1,
"locationId": 67,
"vendorConfig": {
  "vendor": "versa",
  "imageId": 20,
  "productSize": "MEDIUM",
  "directorAddress": "director1.versa.com",
  "controllerAddress": "controller1.versa.com",
  "localAuth": "SDWAN-Branch@Versa.com",
  "remoteAuth": "Controller-1-staging@Versa.com",
  "serialNumber": "Megaport-Hub1"
},
"vnics": [
  {"description": "Data Plane"}
]
}'
```

### JSON from file
```
megaport-cli mve buy --json-file ./mve-config.json
```



## Usage

```
megaport-cli mve buy [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --cost-centre |  |  | Cost centre for billing | false |
| --diversity-zone |  |  | Diversity zone for the MVE | false |
| --interactive | -i | false | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing MVE configuration | false |
| --json-file |  |  | Path to JSON file containing MVE configuration | false |
| --location-id |  | 0 | Location ID where the MVE will be provisioned | false |
| --name |  |  | MVE name | false |
| --promo-code |  |  | Promotional code for discounts | false |
| --resource-tags |  |  | JSON string of key-value resource tags | false |
| --term |  | 0 | Contract term in months (1, 12, 24, or 36) | false |
| --vendor-config |  |  | JSON string containing vendor-specific configuration | false |
| --vnics |  |  | JSON array of network interfaces | false |



