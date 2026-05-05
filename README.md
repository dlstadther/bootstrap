# bootstrap

Personal dotfiles and machine bootstrap managed via symlinks.

## How It Works

`install.sh` walks `dotfiles/` and symlinks each file into `~/` at its matching path. If a file already exists, it's renamed to `<file>.bak.<timestamp>` before the symlink is created. Running it multiple times is safe — already-correct symlinks are skipped.

Machine-specific overrides live in `hosts/<hostname>/` and are applied after the shared dotfiles.

## Setup

**Prerequisites:** [Homebrew](https://brew.sh) installed.

```shell
git clone git@github.com:dlstadther/bootstrap.git ~/code/bootstrap
cd ~/code/bootstrap
./install.sh
make brew-install
```

## Make Targets

| Target | Description |
|---|---|
| `make install` | Symlink all dotfiles into `~/` |
| `make diff` | Show uncommitted changes to dotfiles |
| `make brew-install` | Install packages from `dotfiles/.Brewfile` |
| `make brew-sync` | Show drift between live brew state and `dotfiles/.Brewfile` |
| `make brew-dump` | Write live brew state back to `dotfiles/.Brewfile` |

## Adding a New Machine

1. Create `hosts/<hostname>/` (use `hostname -s` to find the name)
2. Add any machine-specific dotfiles there — they override shared files of the same name
3. Run `./install.sh`

## Structure

```
dotfiles/        # shared configs, mirrors ~/
hosts/
  mbp2022/       # machine-specific overrides
install.sh       # symlink installer
Makefile         # convenience targets
```
