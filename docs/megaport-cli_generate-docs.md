# generate-docs

Generate markdown documentation for the CLI

## Description

Generate comprehensive markdown documentation for the Megaport CLI.

This command will extract all command metadata, examples, and annotations to create a set of markdown files that document the entire CLI interface.

The documentation is organized by command hierarchy, with each command generating its own markdown file containing:
- Command description
- Usage examples
- Available flags
- Subcommands list
- Input/output formats (where applicable)

### Important Notes
  - The output directory will be created if it doesn't exist
  - Existing files in the output directory may be overwritten
  - Hidden commands and 'help' commands are excluded from the documentation

### Example Usage

```
  generate-docs ./docs
```


## Usage

```
megaport-cli generate-docs [directory] [flags]
```







## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--help` | `-h` | `false` | help for generate-docs | false |



