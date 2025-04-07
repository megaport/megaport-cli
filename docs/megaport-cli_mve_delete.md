# delete

Delete an existing MVE

## Description

Delete an existing Megaport Virtual Edge (MVE).

This command allows you to delete an existing MVE by providing its UID.

### Example Usage

```
  delete [mveUID]
  delete [mveUID] --force
  delete [mveUID] --now
```


## Usage

```
megaport-cli mve delete [flags]
```



## Parent Command

* [megaport-cli mve](megaport-cli_mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` | `-f` | `false` | Skip confirmation prompt | false |
| `--now` |  | `false` | Delete resource immediately instead of at end of billing cycle | false |



