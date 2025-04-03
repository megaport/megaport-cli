# update

Update an existing service key

## Description

Update an existing service key for the Megaport API.

This command allows you to modify the details of an existing service key.
You need to specify the key identifier as an argument, and provide any updated values as flags.

Example:
  megaport-cli servicekeys update [key] --description "Updated description"



## Usage

```
megaport-cli servicekeys update [key] [flags]
```

## Examples

```
megaport-cli servicekeys update [key] --description "Updated description"
```

## Parent Command

* [megaport-cli servicekeys](servicekeys.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --active |  | false | Activate the service key | false |
| --description |  |  | Description for the service key | false |
| --product-id |  | 0 | Product ID for the service key | false |
| --product-uid |  |  | Product UID for the service key | false |
| --single-use |  | false | Single-use service key | false |



