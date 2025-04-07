# update

Update an existing service key

## Description

Update an existing service key for the Megaport API.

This command allows you to modify the details of an existing service key. You need to specify the key identifier as an argument, and provide any updated values as flags.

### Example Usage

```
  update a1b2c3d4-e5f6-7890-1234-567890abcdef --description "Updated description"
  update a1b2c3d4-e5f6-7890-1234-567890abcdef --active
  update a1b2c3d4-e5f6-7890-1234-567890abcdef --product-uid "new-product-uid"
```


## Usage

```
megaport-cli servicekeys update [flags]
```



## Parent Command

* [megaport-cli servicekeys](megaport-cli_servicekeys.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--active` |  | `false` | Activate the service key | false |
| `--description` |  |  | Description for the service key | false |
| `--product-id` |  | `0` | Product ID for the service key | false |
| `--product-uid` |  |  | Product UID for the service key | false |
| `--single-use` |  | `false` | Single-use service key | false |



