# update

Update an existing MVE

## Description

Update an existing Megaport Virtual Edge (MVE).

This command allows you to update the details of an existing MVE.
You can provide details in one of three ways:

1. Interactive Mode (default):
 The command will prompt you for each field you can update.

2. Flag Mode:
 Provide the fields you want to update as flags:
   --name, --cost-centre, --contract-term

3. JSON Mode:
 Provide a JSON string or file with the fields you want to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- name: The new name of the MVE.
- cost_centre: The new cost center for the MVE.
- contract_term_months: The new contract term in months (1, 12, 24, or 36).

Example usage:

# Interactive mode (default)
```
megaport-cli mve update [mveUID]
```

# Flag mode
```
megaport-cli mve update [mveUID] --name "New MVE Name" --cost-centre "New Cost Centre" --contract-term 24
```

# JSON mode
```
megaport-cli mve update [mveUID] --json '{"name": "New MVE Name", "costCentre": "New Cost Centre", "contractTermMonths": 24}'
megaport-cli mve update [mveUID] --json-file ./mve-update.json
```



## Usage

```
megaport-cli mve update [mveUID] [flags]
```



## Parent Command

* [megaport-cli mve](mve.md)




## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| --contract-term |  | 0 | New contract term in months (1, 12, 24, or 36) | false |
| --cost-centre |  |  | New cost centre | false |
| --interactive | -i | true | Use interactive mode with prompts | false |
| --json |  |  | JSON string containing MVE update configuration | false |
| --json-file |  |  | Path to JSON file containing MVE update configuration | false |
| --name |  |  | New MVE name | false |



