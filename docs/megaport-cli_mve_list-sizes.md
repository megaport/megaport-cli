# list-sizes

List all available MVE sizes

## Description

List all available MVE sizes from the Megaport API.

This command fetches and displays details about all available MVE instance sizes.
The size you select determines the MVE's capabilities and compute resources.

Each size includes the following specifications:
- `Size`: Size identifier used when creating an MVE (e.g., SMALL, MEDIUM, LARGE)
- `Label`: Human-readable name (e.g., "MVE 2/8", "MVE 4/16")
- `CPU`: Number of virtual CPU cores
- `RAM`: Amount of memory in GB
- `Max CPU Count`: Maximum CPU cores available for the size

Standard MVE sizes available across most vendors:
- `SMALL`: 2 vCPU, 8GB RAM
- `MEDIUM`: 4 vCPU, 16GB RAM
- `LARGE`: 8 vCPU, 32GB RAM
- `X_LARGE_12`: 12 vCPU, 48GB RAM

Note: Not all sizes are available for all vendor images. Some vendors or specific
products may have restrictions on which sizes can be used. Check the image details
using 'megaport-cli mve list-images' for size compatibility.

Example usage:

```
  megaport-cli mve list-sizes
  
```

Example output:
```
  +------------+------------+----------+---------+
  |    SIZE    |   LABEL    |   CPU    |   RAM   |
  +------------+------------+----------+---------+
  | SMALL      | MVE 2/8    | 2 vCPU   | 8 GB    |
  | MEDIUM     | MVE 4/16   | 4 vCPU   | 16 GB   |
  | LARGE      | MVE 8/32   | 8 vCPU   | 32 GB   |
  | X_LARGE_12 | MVE 12/48  | 12 vCPU  | 48 GB   |
  +------------+------------+----------+---------+

```



## Usage

```
megaport-cli mve list-sizes [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)







