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
- `name`: The name of the MVE.
- `term`: The term of the MVE (1, 12, 24, or 36 months).
- `location_id`: The ID of the location where the MVE will be provisioned.
- `vendor-config`: JSON string with vendor-specific configuration (for flag mode)
- `vnics`: JSON array of network interfaces (for flag mode)

Vendor-specific configuration details:
--------------------------------------

6WIND (SixwindVSRConfig):
- `vendor`: Must be "6wind"
- `imageId`: The ID of the 6WIND image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: Custom label for the MVE
- `sshPublicKey`: SSH public key for access

Aruba (ArubaConfig):
- `vendor`: Must be "aruba"
- `imageId`: The ID of the Aruba image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `accountName`: Aruba account name
- `accountKey`: Aruba authentication key
- `systemTag`: System tag for pre-configuration

Aviatrix (AviatrixConfig):
- `vendor`: Must be "aviatrix"
- `imageId`: The ID of the Aviatrix image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: Custom label for the MVE
- `cloudInit`: Cloud-init configuration script

Cisco (CiscoConfig):
- `vendor`: Must be "cisco"
- `imageId`: The ID of the Cisco image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `manageLocally`: Boolean flag to manage locally (true/false)
- `adminSshPublicKey`: Admin SSH public key
- `sshPublicKey`: User SSH public key
- `cloudInit`: Cloud-init configuration script
- `fmcIpAddress`: Firewall Management Center IP address
- `fmcRegistrationKey`: Registration key for FMC
- `fmcNatId`: NAT ID for FMC

Fortinet (FortinetConfig):
- `vendor`: Must be "fortinet"
- `imageId`: The ID of the Fortinet image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `adminSshPublicKey`: Admin SSH public key
- `sshPublicKey`: User SSH public key
- `licenseData`: License data for the Fortinet instance

PaloAlto (PaloAltoConfig):
- `vendor`: Must be "paloalto"
- `imageId`: The ID of the PaloAlto image to use
- `productSize`: (Optional) Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `adminSshPublicKey`: (Optional) Admin SSH public key
- `sshPublicKey`: (Optional) SSH public key for access
- `adminPasswordHash`: (Optional) Hashed admin password
- `licenseData`: (Optional) License data for the PaloAlto instance

Prisma (PrismaConfig):
- `vendor`: Must be "prisma"
- `imageId`: The ID of the Prisma image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: Custom label for the MVE
- `ionKey`: ION key for authentication
- `secretKey`: Secret key for authentication

Versa (VersaConfig):
- `vendor`: Must be "versa"
- `imageId`: The ID of the Versa image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `directorAddress`: Versa director address
- `controllerAddress`: Versa controller address
- `localAuth`: Local authentication string
- `remoteAuth`: Remote authentication string
- `serialNumber`: Serial number for the device

VMware (VmwareConfig):
- `vendor`: Must be "vmware"
- `imageId`: The ID of the VMware image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `adminSshPublicKey`: Admin SSH public key
- `sshPublicKey`: User SSH public key
- `vcoAddress`: VCO address for configuration
- `vcoActivationCode`: Activation code for VCO

Meraki (MerakiConfig):
- `vendor`: Must be "meraki"
- `imageId`: The ID of the Meraki image to use
- `productSize`: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- `mveLabel`: (Optional) Custom label for the MVE
- `token`: Authentication token

Example usage:

### Interactive mode
```
megaport-cli mve buy --interactive

```

### JSON mode - Complete example with full schema
```
megaport-cli mve buy --json '{
  "name": "My MVE Display Name",
  "term": 12,
  "locationId": 123,
  "diversityZone": "zone-1",
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
    "cloudInit": "#cloud-config\npackages:\n - nginx\n",
    "fmcIpAddress": "10.0.0.1",
    "fmcRegistrationKey": "key123",
    "fmcNatId": "natid123"
  },
  "vnics": [
    {"description": "Data Plane", "vlan": 100},
    {"description": "Management", "vlan": 200}
  ]
}'

```

Notes:
- For production deployments, you may want to use a JSON file to manage complex configurations
- `To list available images and their IDs, use`: megaport-cli mve list-images
- `To list available sizes, use`: megaport-cli mve list-sizes
- `Location IDs can be retrieved with`: megaport-cli locations list



## Usage

```
megaport-cli mve buy [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--cost-centre` |  |  | Cost centre for billing | false |
| `--diversity-zone` |  |  | Diversity zone for the MVE | false |
| `--interactive` | `-i` | `false` | Use interactive mode with prompts | false |
| `--json` |  |  | JSON string containing MVE configuration | false |
| `--json-file` |  |  | Path to JSON file containing MVE configuration | false |
| `--location-id` |  | `0` | Location ID where the MVE will be provisioned | false |
| `--name` |  |  | MVE name | false |
| `--promo-code` |  |  | Promotional code for discounts | false |
| `--term` |  | `0` | Contract term in months (1, 12, 24, or 36) | false |
| `--vendor-config` |  |  | JSON string containing vendor-specific configuration | false |
| `--vnics` |  |  | JSON array of network interfaces | false |



