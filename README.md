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

## Daily Use

### tmux

**Session shortcuts** (defined in `.zsh/tmux.zsh`):

| Alias | Command | Description |
|---|---|---|
| `t` | `tmux attach` | Attach to most recent session |
| `ta <name>` | `tmux attach -t <name>` | Attach to named session |
| `tl` | `tmux list-sessions` | List all sessions |
| `tn <name>` | `tmux new-session -s <name>` | Create named session |
| `tk <name>` | `tmux kill-session -t <name>` | Kill named session |
| `td` | `tmux detach` | Detach from current session |

**Key bindings** (prefix = `Ctrl+b`):

| Binding | Action |
|---|---|
| `prefix + \|` | Split pane horizontally |
| `prefix + -` | Split pane vertically |
| `prefix + h/j/k/l` | Navigate panes (vim-style) |
| `prefix + H/J/K/L` | Resize panes |
| `prefix + o` | Fuzzy session picker (tmux-sessionx) |
| `prefix + C-c` | New session |
| `prefix + x` | Kill current pane |
| `prefix + r` | Reload tmux config |

**Session persistence** (tmux-resurrect + tmux-continuum):

Sessions auto-save every 15 minutes. On a fresh tmux start, the last session restores automatically. To force a restore manually:

```shell
tmux-restore        # from inside a running tmux session
# or: prefix + Ctrl+r
```

**Configured sessions** (`~/.config/tmux/bootstrap.sh`):

`tmux-bootstrap` reads `~/.config/tmux/bootstrap.sh` to create your standard named sessions. This file is not tracked in the repo. After first install, create it and define your sessions:

```shell
# One-time setup: create personal session bootstrap
cp ~/.config/tmux/bootstrap.example.sh ~/.config/tmux/bootstrap.sh
$EDITOR ~/.config/tmux/bootstrap.sh

# Then on each login (or after a reboot):
tmux new-session -d -s main   # start tmux if not running
tmux-start                    # restore sessions + run bootstrap
```

---

### tmux-notify

Sends a macOS notification when a long-running command finishes. The trigger threshold is 5 seconds (set via `@tnotify-sleep-duration` in `tmux.conf`).

- Clicking the notification focuses the correct Ghostty window and switches to the tmux session/window/pane where the command ran
- Falls back to `osascript` if `terminal-notifier` is unavailable
- The patch (`scripts/patch-tmux-notify.sh`) is applied automatically by `make init-tmux` and must be re-applied after plugin updates (`prefix + U` → `make init-tmux`)

No configuration needed day-to-day — it activates automatically for any command that runs longer than the threshold.

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
