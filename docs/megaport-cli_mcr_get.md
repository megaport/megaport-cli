# get

Get details for a single MCR

## Description

Get details for a single MCR.

This command retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
You must provide the unique identifier (UID) of the MCR you wish to retrieve.

The output includes:
- UID: Unique identifier of the MCR
- Name: User-defined name of the MCR
- Location ID: Physical location where the MCR is provisioned
- Port Speed: Speed of the MCR (e.g., 1000, 2500, 5000, 10000 Mbps)
- Provisioning Status: Current provisioning status of the MCR (e.g., Active, Inactive, Deleting)

Example usage:
```
megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef

```


## Usage

```
megaport-cli mcr get [mcrUID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|



