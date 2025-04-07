# completion

Generate completion script

## Description

Generate shell completion scripts for Megaport CLI.

This command outputs shell completion code for various shell environments that can be used to enable tab-completion of Megaport CLI commands.

Important notes:
  - Bash: source <(megaport-cli completion bash)
  - Zsh: You need to enable shell completion with 'autoload -U compinit; compinit'
  - Fish: megaport-cli completion fish | source
  - PowerShell: megaport-cli completion powershell | Out-String | Invoke-Expression

Example usage:

```
  completion bash > ~/.bash_completion.d/megaport-cli
  completion zsh > "${fpath[1]}/_megaport-cli"
  completion fish > ~/.config/fish/completions/megaport-cli.fish
  completion powershell > megaport-cli.ps1
```


## Usage

```
megaport-cli completion [bash|zsh|fish|powershell] [flags]
```







## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|



