# Dotfiles Redesign — Symlink-Based Config Management

**Date:** 2026-05-04
**Status:** Approved

## Problem

Config drift accumulates because updating a machine requires remembering to also update the repo. The Ansible-based approach adds friction: the deploy workflow is unfamiliar, the `system/` and `ansible/` directories have diverged, and the mental model of "repo → machine" is backwards from how changes actually happen (machine first, repo second).

## Goal

Eliminate drift structurally. The repo files become the live files via symlinks — editing `~/.zshrc` is editing the repo file. No sync step exists because no sync is needed.

## Scope

**In scope:**
- Zsh configuration (`.zshrc`, `.zsh/`, `.zsh.before/`, `.zsh.after/`)
- Vim configuration (`.vimrc`)
- Homebrew Brewfile (`~/.Brewfile`)
- App configs (e.g., `~/.config/ghostty/`)
- Per-machine overrides (mbp2022 initially)

**Out of scope:**
- Package installation (Homebrew, apt, vim, zsh) — tools must already exist
- macOS system preferences / `defaults write`
- Secrets — `dotfiles/.zsh/secrets.zsh` sources from `~/.secrets` (not in repo)
- Linux support (lemp9) — removed for now; can be re-added under `hosts/lemp9/`

## Repository Structure

```
bootstrap/
├── install.sh              # symlink installer; detects machine via hostname -s
├── Makefile                # make install, diff, brew-install, brew-sync, brew-dump
├── README.md
├── CLAUDE.md
├── dotfiles/               # shared dotfiles; mirrors ~/
│   ├── .zshrc
│   ├── .zsh/
│   │   ├── aliases.zsh
│   │   ├── prompt.zsh
│   │   └── ...
│   ├── .zsh.before/
│   │   └── 0000_reset.zsh
│   ├── .zsh.after/
│   │   └── zzzz_placeholder.zsh
│   ├── .vimrc
│   ├── .Brewfile
│   └── .config/
│       └── ghostty/        # placeholder; populate when ghostty config is ready
├── hosts/
│   └── mbp2022/            # machine-specific overrides; same relative paths as dotfiles/
└── keyboardio/             # unchanged
```

## install.sh Behavior

### Machine Detection
Hostname detected via `hostname -s`. If `hosts/<machine>/` does not exist, only shared dotfiles are installed — new machines work without any host directory.

### Symlink Granularity
The install script symlinks **individual files**, not directories. This prevents tools from writing generated files into the repo (e.g., ghostty writing a cache file next to its config). Parent directories (e.g., `~/.config/ghostty/`) are created if they don't exist, but are not themselves symlinked.

### Symlink Logic (per file)
1. Target does not exist → create parent directories + create symlink
2. Target is already a symlink pointing into this repo → skip (idempotent)
3. Target is an existing file or foreign symlink → rename to `<target>.bak.<timestamp>`, then create symlink

### Order of Operations
1. Shared files from `dotfiles/` symlinked first
2. Machine-specific files from `hosts/<machine>/` symlinked second, overriding shared files of the same name

### What install.sh Does NOT Do
- Does not install any software
- Does not run `brew bundle install`
- Does not modify anything outside `~/`
- Does not touch system preferences

## Makefile Targets

| Target | Command | Notes |
|---|---|---|
| `make install` | `./install.sh` | Create symlinks; backup conflicts |
| `make diff` | `git diff` | Live files are repo files — this is the diff |
| `make brew-install` | `brew bundle install --global` | Install from `~/.Brewfile` (repo symlink) |
| `make brew-sync` | Dump to `/tmp/.Brewfile.current`, diff vs `dotfiles/.Brewfile` | Read-only; shows bidirectional drift |
| `make brew-dump` | `brew bundle dump --force` | Writes live state to `~/.Brewfile` → repo file |

### Brewfile Drift Workflow
1. `make brew-sync` — see what's drifted (read-only)
2. `make brew-dump` — pull live state into repo Brewfile
3. `git diff` — review; prune experimental installs
4. `git commit`

## Per-Machine Overrides

`hosts/<machine>/` uses the same relative path structure as `dotfiles/`. A file at `hosts/mbp2022/.zsh/aliases.zsh` is symlinked to `~/.zsh/aliases.zsh`, overriding the shared version.

`hosts/mbp2022/` starts empty. It exists as a placeholder for future divergence (e.g., work-specific aliases, machine-specific PATH entries).

## Migration

| Removed | Reason |
|---|---|
| `ansible/` | Replaced by symlink approach |
| `system/` | Vestigial; diverged from ansible; never authoritative |
| `apps.md` | Stale (references Atom, YADR, etc.) |

Files move from `ansible/roles/workstations/files/` to `dotfiles/`:
- `zsh/.zshrc` → `dotfiles/.zshrc`
- `zsh/.zsh/*.zsh` → `dotfiles/.zsh/*.zsh`
- `zsh/.zsh.before/` → `dotfiles/.zsh.before/`
- `zsh/.zsh.after/` → `dotfiles/.zsh.after/`
- `vim/.vimrc` → `dotfiles/.vimrc`
- `homebrew/.Brewfile-mac` → `dotfiles/.Brewfile`

The `keyboardio/` directory is unchanged.

## Prerequisites (New Machine Setup)

Before running `make install`:
- Zsh installed and set as default shell
- Git installed (to clone the repo)
- Homebrew installed (if brew targets are needed)
- Vim installed (if `.vimrc` should have effect)

## Caveats

- `dotfiles/.zsh/secrets.zsh` must source from `~/.secrets` (outside the repo) — never commit credentials
- Linux support is intentionally removed; re-add under `hosts/lemp9/` when needed
- The Brewfile in this repo is macOS-only; a `hosts/lemp9/` future addition would need its own Brewfile handling
