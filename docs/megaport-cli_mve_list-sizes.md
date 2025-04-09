# list-sizes

List all available MVE sizes

## Description

List all available MVE sizes from the Megaport API.

This command fetches and displays details about all available MVE instance sizes. The size you select determines the MVE's capabilities and compute resources.

### Important Notes
  - Standard MVE sizes available across most vendors: SMALL (2 vCPU, 8GB RAM), MEDIUM (4 vCPU, 16GB RAM), LARGE (8 vCPU, 32GB RAM), X_LARGE_12 (12 vCPU, 48GB RAM)
  - Not all sizes are available for all vendor images. Check the image details using 'megaport-cli mve list-images' for size compatibility

### Example Usage

```sh
  list-sizes
```


## Usage

```sh
megaport-cli mve list-sizes [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|



