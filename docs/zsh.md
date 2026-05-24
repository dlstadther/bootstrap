# Zsh Configuration

Zsh config lives in [`dotfiles/`](../dotfiles/) and is symlinked into `~/` by `make install`.

## Entry Point

[`.zshrc`](../dotfiles/.zshrc) sources every file matching `~/.zsh/*.zsh` in alphabetical order. That's it — one line, no manual registration required.

```zsh
for config_file ($HOME/.zsh/*.zsh) source $config_file
```

## `~/.zsh/` — Per-Tool Config Files

Each tool or concern gets its own file. Files in this repo's [`dotfiles/.zsh/`](../dotfiles/.zsh/):

| File | Purpose |
|---|---|
| `0_path.zsh` | PATH setup (runs first due to `0_` prefix) |
| `0000_before.zsh` | Early setup (XDG, lang, editor) |
| `aliases.zsh` | Shell, git, and AI aliases |
| `completions.zsh` | Completion system init |
| `gcloud.zsh` | Google Cloud SDK |
| `golang.zsh` | Go env and PATH |
| `mise.zsh` | mise version manager |
| `nvm.zsh` | Node version manager |
| `options_history.zsh` | History options |
| `options_other.zsh` | Misc zsh options |
| `pnpm.zsh` | pnpm PATH |
| `prek.zsh` | Pre-exec hooks |
| `prompt.zsh` | Prompt (Starship) |
| `pyenv.zsh` | Python version manager |
| `rbenv.zsh` | Ruby version manager |
| `reflex.zsh` | reflex file watcher |
| `rm.zsh` | Safe rm (trash instead of delete) |
| `secrets.zsh` | Sources `~/.secrets` if it exists — never committed |
| `tmux.zsh` | tmux aliases and session shortcuts |
| `vi-mode.zsh` | Vi key bindings in shell |
| `zzzz_after.zsh` | Late-running setup (runs last due to `zzzz_` prefix) |

### Loading Order

Alphabetical, so numeric/word prefixes control order when it matters:
- `0_path.zsh` runs before `aliases.zsh`
- `zzzz_after.zsh` runs last

## Hook Directories

### `.zsh.before/`

Files sourced **before** `~/.zsh/`. Used for setup that must happen before any tool config loads (e.g., resetting PATH to a clean state).

- [`0000_reset.zsh`](../dotfiles/.zsh.before/0000_reset.zsh) — resets PATH to system default before the main configs build it up

> `.zshrc` does not source `.zsh.before/` by default — this depends on your prompt framework or a hook in `0000_before.zsh`. Adjust if you change the entry point.

### `.zsh.after/`

Files sourced **after** `~/.zsh/`. Placeholder for post-config overrides.

- [`zzzz_placeholder.zsh`](../dotfiles/.zsh.after/zzzz_placeholder.zsh) — empty placeholder

## Extension: Local, Untracked Files

Because `.zshrc` sources **all** `~/.zsh/*.zsh` files, you can drop any `.zsh` file into `~/.zsh/` and it will be picked up automatically — even if it's not in this repo.

This is intentional. Use it for:
- Machine-local scripts and functions that shouldn't be committed
- Work credentials or environment setup
- Personal aliases that don't belong in the shared config

**Example:** `~/.zsh/scripts.zsh` — a local file for one-off shell functions, sourced automatically on every new shell, never committed to this repo.

The same applies to `~/.secrets` — sourced by `secrets.zsh` if it exists, but never tracked.

## Machine-Specific Overrides

Files in `hosts/<hostname>/` use the same relative path structure as `dotfiles/` and are symlinked second, overriding any shared file of the same name.

See [`hosts/`](../hosts/) for existing machine configs.
