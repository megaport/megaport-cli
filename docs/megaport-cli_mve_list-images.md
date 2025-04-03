# list-images

List all available MVE images

## Description

List all available MVE images from the Megaport API.

This command fetches and displays a list of all available MVE images with details
about each one. These images are used when creating new MVEs with the 'buy' command.

The output includes:
- `ID`: Unique identifier required for the 'buy' command
- `Vendor`: The network function vendor (e.g., Cisco, Fortinet, Palo Alto)
- `Product`: Specific product name (e.g., C8000, FortiGate-VM, VM-Series)
- `Version`: Software version of the image
- `Release`: Whether this is a production release image (true) or development/beta (false)
- `Sizes`: Available instance sizes (SMALL, MEDIUM, LARGE, X_LARGE_12)
- `Description`: Additional vendor-specific information when available

Available filters:
  --vendor string        Filter images by vendor name (e.g., "Cisco", "Fortinet")
  --product-code string  Filter images by product code
  --id int               Filter images by exact image ID
  --version string       Filter images by version string
  --release-image        Only show official release images (excludes beta/development)

Example usage:

### List all available images
```
  megaport-cli mve list-images

```

### List only Cisco images
```
  megaport-cli mve list-images --vendor "Cisco"

```

### List only release (production) images for Fortinet
```
  megaport-cli mve list-images --vendor "Fortinet" --release-image

```

Example output:
```
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+
  | ID  |  VENDOR  |            PRODUCT               |   VERSION    | RELEAS |        SIZES         |      DESCRIPTION        |
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+
  | 83  | Cisco    | C8000                            | 17.15.01a    | true  | SMALL,MEDIUM,LARGE   |                         |
  | 78  | Cisco    | Secure Firewall Threat Defense   | 7.4.2-172    | true  | MEDIUM,LARGE         |                         |
  | 57  | Fortinet | FortiGate-VM                     | 7.0.14       | true  | SMALL,MEDIUM,LARGE   |                         |
  | 65  | Palo Alto| VM-Series                        | 10.2.9-h1    | true  | SMALL,MEDIUM,LARGE   |                         |
  | 88  | Palo Alto| Prisma SD-WAN 310xv              | vION 3102v-  | true  | SMALL                | Requires MVE Size 2/8   |
  | 62  | Meraki   | vMX                              | 20231214     | false | SMALL,MEDIUM,LARGE   | Engineering Build - Not |
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+

```



## Usage

```
megaport-cli mve list-images [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--id` |  | `0` | Filter images by ID | false |
| `--product-code` |  |  | Filter images by product code | false |
| `--release-image` |  | `false` | Filter images by release image | false |
| `--vendor` |  |  | Filter images by vendor | false |
| `--version` |  |  | Filter images by version | false |



