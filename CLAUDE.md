# CLAUDE.md

## What This Is

A personal system bootstrap and configuration automation project using Ansible. It automates developer machine setup (macOS MBP 2022, Linux LEMP9 server) including Zsh configuration, Homebrew packages, Vim, and JetBrains Toolbox.

## Commands

All Ansible commands run from the `ansible/` directory with uv.

**Setup (one-time):**
```shell
make build
```

**Run playbooks:**
```shell
# Dry run (check mode)
uv run ansible-playbook mbp2022.yml --ask-become-pass --check

# Full macOS setup
uv run ansible-playbook mbp2022.yml --ask-become-pass

# Selective tags (e.g., only zsh)
uv run ansible-playbook mbp2022.yml --ask-become-pass --tags "zsh"

# LEMP9 server
uv run ansible-playbook lemp9.yml --ask-become-pass --tags "apt,zsh"

# Test connectivity
uv run ansible-playbook test.yml --ask-become-pass
```

**Quick Zsh-only copy (no Ansible):**
```shell
cp -r ./ansible/roles/workstations/files/zsh/ ~/
```

## Architecture

### Ansible Structure

- `ansible/mbp2022.yml` — macOS playbook (runs `workstations` role)
- `ansible/lemp9.yml` — Linux server playbook (runs `common` + `workstations` roles)
- `ansible/roles/workstations/tasks/` — Task files imported conditionally by OS
- `ansible/roles/workstations/files/` — Dotfiles and config managed by Ansible

Tasks are tagged (e.g., `homebrew`, `zsh`, `vim`, `apt`) for selective execution. OS detection via `ansible_distribution` drives conditional task inclusion.

### Zsh Configuration

Located in `ansible/roles/workstations/files/zsh/`:
- `.zshrc` — Entry point, sources files from `.zsh/`
- `.zsh/` — Modular configs loaded in order (prefixed numerically: `0_path.zsh`, etc.)
- `.zsh.before/` / `.zsh.after/` — Hook directories for pre/post configs

Each tool gets its own file (e.g., `pyenv.zsh`, `nvm.zsh`, `golang.zsh`, `aliases.zsh`).

### Brewfiles

- `ansible/roles/workstations/files/homebrew/.Brewfile-mac` — macOS packages
- `ansible/roles/workstations/files/homebrew/.Brewfile-linux` — LinuxBrew packages

Ansible copies the appropriate Brewfile to `~/.Brewfile` based on detected OS.

### Inventory

`ansible/inventory` defines three local hosts: `localhost`, `mbp2022`, `lemp9`. All use `ansible_connection=local`.
