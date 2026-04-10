# generate-docs

Generate documentation for the CLI

## Description

Generate documentation for the Megaport CLI.

By default (--format markdown) this command extracts all command metadata, examples, and annotations to create a set of markdown files that document the entire CLI interface.

Use --format man to generate Unix man pages for all commands instead.

The documentation is organized by command hierarchy, with each command generating its own file containing:
- Command description
- Usage examples
- Available flags
- Subcommands list
- Input/output formats (where applicable)

### Important Notes
  - The output directory will be created if it doesn't exist
  - Existing files in the output directory may be overwritten
  - Hidden commands are excluded from both formats
  - The help command is excluded from markdown output; man format includes it via cobra/doc
  - Man pages can be viewed with: man <outputDir>/megaport-cli.1

### Example Usage

```sh
  megaport-cli generate-docs ./docs
  megaport-cli generate-docs --format man ./man/
```

## Usage

```sh
megaport-cli generate-docs [flags]
```


## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|
| `--format` |  | `markdown` | Output format: markdown or man | false |
| `--help` | `-h` | `false` | help for generate-docs | false |

## Subcommands
* [docs](megaport-cli_generate-docs_docs.md)

