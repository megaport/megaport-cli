# list-images

List all available MVE images

## Description

List all available MVE images from the Megaport API.

This command fetches and displays a list of all available MVE images with details about each one. These images are used when creating new MVEs with the 'buy' command.

Optional fields:
release-image: Only show official release images (excludes beta/development)
vendor: Filter images by vendor name (e.g., "Cisco", "Fortinet")
product-code: Filter images by product code
id: Filter images by exact image ID
version: Filter images by version string

Important notes:
- The output includes the image ID, vendor, product, version, release status, available sizes, and description
- The ID field is required when specifying an image in the 'buy' command

Example usage:

list-images
list-images --vendor "Cisco"
list-images --vendor "Fortinet" --release-image



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



