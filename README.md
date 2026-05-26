# bootstrap

Personal dotfiles and machine bootstrap, managed via symlinks. Editing a config file in `~/` is editing this repo — no sync step.

## What's Configured

- **Zsh** — modular per-tool configs, aliases, prompt, vi-mode ([details](docs/zsh.md))
- **tmux** — keybindings, session aliases, plugins, session bootstrap ([details](docs/tmux.md))
- **Claude Code** — settings, rules, and status line config (`dotfiles/.claude/`)
- **Git** — global gitconfig (`dotfiles/.gitconfig`)
- **Ghostty** — terminal config (`dotfiles/.config/ghostty/`)
- **Vim** — vimrc (`dotfiles/.vimrc`)
- **Homebrew** — shared Brewfile (`dotfiles/.Brewfile`)

## Fresh Machine Setup

**Prerequisite:** [Homebrew](https://brew.sh) installed.

```shell
git clone git@github.com:dlstadther/bootstrap.git ~/code/bootstrap
cd ~/code/bootstrap
make bootstrap        # symlink dotfiles + apply macOS defaults
make brew-install     # install packages from dotfiles/.Brewfile
make init-tmux        # install TPM + all tmux plugins
```

## Updating an Existing Machine

```shell
git pull --rebase
make install          # rebuild bs binary + pick up new dotfiles
bs audit              # review dotfile symlink and brew drift
```

## bs CLI

`bs` is the unified CLI for ongoing machine and dotfile management. Built and installed via `make install`.

```shell
bs help               # list all commands
bs version            # show installed vs repo commit hash
bs audit              # check dotfile symlinks + brew drift
bs brew sync          # show drift between live brew state and .Brewfile
bs brew dump          # write live brew state back to .Brewfile
bs brew install       # install packages from .Brewfile
bs tmux add --cwd <path> [--name <name>] [--agent claude]
                      # open an agent workspace in tmux
```

## Scripts

Setup automation lives in [`scripts/`](scripts/):

| Script | Purpose |
|---|---|
| `install.sh` | Idempotent symlink installer — backs up conflicts, skips correct symlinks |
| `macos-defaults.sh` | Apply macOS system defaults (Finder, Dock, keyboard) |
| `init-tmux.sh` | Install/update TPM and plugins, apply patches, reload config |

## Make Targets

| Target | Description |
|---|---|
| `make install` | Build `bs`, then symlink all dotfiles into `~/` |
| `make bootstrap` | Full setup: dotfiles + macOS defaults |
| `make init-tmux` | Install/update TPM + plugins |
| `make brew-install` | Install packages from `dotfiles/.Brewfile` |
| `make brew-sync` | Show drift between live brew state and `dotfiles/.Brewfile` |
| `make brew-dump` | Write live brew state back to `dotfiles/.Brewfile` |
| `make build-bs` | Build the `bs` binary only |
| `make test-bs` | Run `bs` CLI unit tests |

See the [`Makefile`](Makefile) for all targets.

## Adding a New Machine

1. Find the hostname: `hostname -s`
2. Create `hosts/<hostname>/`
3. Add machine-specific dotfiles there — same relative paths as `dotfiles/`, applied after shared files
4. Run `make install`

## Structure

```
dotfiles/        # shared configs, mirrors ~/
hosts/
  <hostname>/    # machine-specific overrides (detected via hostname -s)
scripts/         # install, macos-defaults, init-tmux
cli/             # bs CLI source (Go)
docs/            # detailed documentation
Makefile
```
