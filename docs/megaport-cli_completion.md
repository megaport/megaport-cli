# completion

Generate completion script

## Description

### Important Notes
  - Bash: source <(megaport-cli completion bash)
  - Zsh: You need to enable shell completion with 'autoload -U compinit; compinit'
  - Fish: megaport-cli completion fish | source
  - PowerShell: megaport-cli completion powershell | Out-String | Invoke-Expression

### Example Usage

```
  completion bash > ~/.bash_completion.d/megaport-cli
  completion zsh > "${fpath[1]}/_megaport-cli"
  completion fish > ~/.config/fish/completions/megaport-cli.fish
  completion powershell > megaport-cli.ps1
```


## Usage

```
megaport-cli completion [flags]
```







## Flags

| Name | Shorthand | Default | Description | Required |
|------|-----------|---------|-------------|----------|



