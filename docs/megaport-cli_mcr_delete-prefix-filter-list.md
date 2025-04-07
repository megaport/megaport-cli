# delete-prefix-filter-list

Delete a prefix filter list on an MCR

## Description

Delete a prefix filter list on an MCR.

This command allows you to delete a prefix filter list from the specified MCR.

### Optional Fields
  - `force`: Force deletion without confirmation

### Example Usage

```
  delete-prefix-filter-list [mcrUID] [prefixFilterListID]
  delete-prefix-filter-list [mcrUID] [prefixFilterListID] --force
```


## Usage

```
megaport-cli mcr delete-prefix-filter-list [mcrUID] [prefixFilterListID] [flags]
```



## Parent Command

* [megaport-cli mcr](megaport-cli_mcr.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--force` |  | `false` | Force deletion without confirmation | false |



