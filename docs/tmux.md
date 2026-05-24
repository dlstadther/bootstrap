# tmux Configuration

tmux config lives in [`dotfiles/.config/tmux/`](../dotfiles/.config/tmux/) and is symlinked into `~/.config/tmux/` by `make install`.

## Setup

Install or update TPM and all plugins:

```shell
make init-tmux
```

Run this after first install and after any changes to `.tmux.conf` (especially after `prefix + U` to update plugins, since patches must be re-applied).

## Key Bindings

Prefix is `Ctrl+b`.

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
| `prefix + Ctrl+s` | Save session (tmux-resurrect) |
| `prefix + Ctrl+r` | Restore session (tmux-resurrect) |

## Session Aliases

Defined in [`dotfiles/.zsh/tmux.zsh`](../dotfiles/.zsh/tmux.zsh):

| Alias | Command | Description |
|---|---|---|
| `t` | `tmux attach` | Attach to most recent session |
| `ta <name>` | `tmux attach -t <name>` | Attach to named session |
| `tl` | `tmux list-sessions` | List all sessions |
| `tn <name>` | `tmux new-session -s <name>` | Create named session |
| `tk <name>` | `tmux kill-session -t <name>` | Kill named session |
| `td` | `tmux detach` | Detach from current session |

## Plugins

Managed via [TPM](https://github.com/tmux-plugins/tpm).

| Plugin | Purpose |
|---|---|
| tmux-resurrect | Manual session save/restore |
| tmux-continuum | Auto-saves sessions every 15 min; auto-restores on tmux start |
| tmux-sessionx | Fuzzy session picker (`prefix + o`) |
| tmux-notify | macOS notification when a long command finishes |

### tmux-notify

Fires a notification when a command runs longer than 5 seconds (configured via `@tnotify-sleep-duration`). Clicking the notification focuses the correct Ghostty window and switches to the session/window/pane where the command ran.

Requires a local patch (`scripts/patch-tmux-notify.sh`) to replace the default `osascript` call with `terminal-notifier`. Applied automatically by `make init-tmux` — re-run it after any plugin update.

## Session Bootstrap

`tmux-bootstrap` reads `~/.config/tmux/bootstrap.sh` to create your standard named sessions on login. This file is **not tracked in the repo**.

First-time setup:

```shell
cp ~/.config/tmux/bootstrap.example.sh ~/.config/tmux/bootstrap.sh
$EDITOR ~/.config/tmux/bootstrap.sh
```

To start sessions:

```shell
tmux-start    # restore saved sessions + run bootstrap
```

See [`hosts/`](../hosts/) for machine-specific session templates (e.g., `workspace.sh`).
