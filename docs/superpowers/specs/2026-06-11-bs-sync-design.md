# bs sync — Unified Machine Sync Command

**Issue:** bs-ci7  
**Date:** 2026-06-11

## Problem

Bootstrap requires 4 manual steps in implicit order to keep a machine in sync:

1. `make install` — build `bs` binary + symlink dotfiles
2. `mise install` — install tool versions
3. `make brew-install` — install Homebrew packages
4. `make install-plugins` — install Claude plugins

The `brew bundle install` step is slow even when nothing has changed. There is no single command for the common daily-use case.

## Design

### Command Surface

| Command | Does | When to use |
|---|---|---|
| `make install` | Build `bs` binary + symlink dotfiles | Dotfile or CLI source changes |
| `bs sync` | Sync runtime state: mise, brew, plugins | Daily sync |
| `make sync` | Both of the above in sequence | Full resync from scratch |

### `bs sync`

New top-level Cobra command registered in `cli/cmd/root.go`. Implementation in `cli/cmd/sync.go`.

Runs three steps in order:

**1. mise**
- Shells out to `mise install`
- Already idempotent; no pre-check needed

**2. brew**
- Runs `brew bundle check --global` first
- Only runs `brew bundle install --global` if the check exits non-zero
- `--force` flag skips the check and runs install unconditionally

**3. plugins**
- Reads `~/.claude/settings.json` via `jq`
- Installs each plugin where `enabledPlugins[name] == true` via `claude plugin install <name>`
- Mirrors the current `scripts/install-plugins.sh` behavior (no marketplace registration)

**Error handling:** each step is independent. A failure prints the error and continues so one broken step does not abort the rest. Exit code is non-zero if any step failed.

**Output:** each step prints a single status line before running (e.g., `syncing mise...`, `syncing brew...`, `syncing plugins...`).

### `make sync` target

```makefile
sync: install
	bs sync
```

Depends on `install` so the `bs` binary is always rebuilt before `bs sync` runs.

### What doesn't change

- `make install`, `make bootstrap`, `make brew-install`, `make install-plugins` — all unchanged
- `scripts/install-plugins.sh` — kept as-is; still used by `make install-plugins`

## Files to touch

- `cli/cmd/sync.go` — new file, `bs sync` command
- `cli/cmd/root.go` — register sync command
- `Makefile` — add `sync` target, add `sync` to `.PHONY`
