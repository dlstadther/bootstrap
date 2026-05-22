# bootstrap

Personal dotfiles and machine bootstrap managed via symlinks.

## How It Works

`scripts/install.sh` walks `dotfiles/` and symlinks each file into `~/` at its matching path. If a file already exists it is renamed to `<file>.bak.<timestamp>` before the symlink is created. Running it multiple times is safe — already-correct symlinks are skipped.

Machine-specific overrides live in `hosts/<hostname>/` and are applied after the shared dotfiles.

---

## Fresh Machine Setup

**Prerequisites:** [Homebrew](https://brew.sh) installed.

```shell
git clone git@github.com:dlstadther/bootstrap.git ~/code/bootstrap
cd ~/code/bootstrap
make bootstrap        # symlink dotfiles + apply macOS defaults
make brew-install     # install packages from dotfiles/.Brewfile
```

### Post-install steps

**tmux — install/update TPM:**

```shell
make init-tmux
```

Installs/updates TPM, installs and updates all plugins, applies local patches, and reloads the tmux config if a server is running.

---

## Updating an Existing Machine

Pull the latest changes and re-run the installer. The installer is idempotent — it skips symlinks that are already correct and only touches new or changed files.

```shell
cd ~/code/bootstrap
git pull --rebase
make install          # pick up any new dotfiles
make brew-sync        # review drift between live brew state and dotfiles/.Brewfile
make brew-install     # install any newly added packages
```

To install or update tmux plugins after changes to `.tmux.conf`:

```shell
make init-tmux
```

---

## Make Targets

| Target | Description |
|---|---|
| `make install` | Symlink all dotfiles into `~/` |
| `make bootstrap` | Full machine setup: dotfiles + macOS defaults |
| `make macos-defaults` | Apply macOS system defaults (Finder, Dock, keyboard, etc.) |
| `make init-tmux` | Install/update TPM + all plugins, apply patches, reload config |
| `make diff` | Show uncommitted changes to dotfiles |
| `make brew-install` | Install packages from `dotfiles/.Brewfile` |
| `make brew-sync` | Show drift between live brew state and `dotfiles/.Brewfile` |
| `make brew-dump` | Write live brew state back to `dotfiles/.Brewfile` |

---

## Adding a New Machine

1. Find the hostname: `hostname -s`
2. Create `hosts/<hostname>/`
3. Add machine-specific dotfiles there — they override shared files of the same name
4. Run `make install`

---

## Structure

```
dotfiles/        # shared configs, mirrors ~/
hosts/
  <hostname>/    # machine-specific overrides (detected via hostname -s)
scripts/
  install.sh     # idempotent symlink installer
  macos-defaults.sh  # macOS system preference defaults
  init-tmux.sh   # TPM + plugin install/update + local patches
  patch-tmux-notify.sh  # replaces osascript with terminal-notifier in tmux-notify
Makefile         # convenience targets
```
