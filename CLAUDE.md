# CLAUDE.md

## What This Is

A personal system bootstrap and configuration automation project using Ansible. Automates developer machine setup (macOS MBP 2022, Linux LEMP9 server) including Zsh configuration, Homebrew packages, Vim, and JetBrains Toolbox.

## Commands

All Ansible commands run from the `ansible/` directory.

**Setup (one-time):**
```shell
mise install   # installs uv via mise
uv sync        # installs ansible and dependencies
```

**Run playbooks:**
```shell
# Dry run (check mode)
uv run ansible-playbook mbp2022.yml --ask-become-pass --check

# Full macOS setup
uv run ansible-playbook mbp2022.yml --ask-become-pass

# Selective tags
uv run ansible-playbook mbp2022.yml --ask-become-pass --tags "zsh"

# LEMP9 server
uv run ansible-playbook lemp9.yml --ask-become-pass --tags "apt,zsh"

# Test connectivity
uv run ansible-playbook test.yml --ask-become-pass

# Gather host facts
uv run ansible all -m setup
```

**Quick Zsh-only copy (no Ansible):**
```shell
cp -r ./ansible/roles/workstations/files/zsh/ ~/
```

**Available tags:** `vim`, `apt`, `jetbrains`, `zsh`, `homebrew`

## Architecture

### Ansible Structure

- `ansible/mbp2022.yml` — macOS playbook (runs `workstations` role)
- `ansible/lemp9.yml` — Linux server playbook (runs `common` + `workstations` roles)
- `ansible/roles/workstations/tasks/main.yml` — imports task files conditionally by OS and tag
- `ansible/roles/workstations/files/` — dotfiles and configs deployed by Ansible

OS detection via `ansible_distribution` drives conditional task inclusion (e.g., `apt` and `jetbrains` tasks only run on Debian/Ubuntu/Pop!_OS; `homebrew` zsh install only on MacOSX).

### Inventory

`ansible/inventory` defines hosts `localhost`, `mbp2022`, `lemp9` under the `[python3]` group. All use `ansible_connection=local`. `mbp2022.yml` uses `.venv/bin/python` as the interpreter; `lemp9.yml` uses `/usr/bin/python3`.

### Zsh Configuration

Located in `ansible/roles/workstations/files/zsh/`:
- `.zshrc` — entry point, sources files from `.zsh/`
- `.zsh/` — modular configs loaded in numeric order (`0_path.zsh`, `0000_before.zsh`, etc.)
- `.zsh.before/` / `.zsh.after/` — hook directories for pre/post configs

Each tool gets its own file (e.g., `pyenv.zsh`, `nvm.zsh`, `golang.zsh`, `aliases.zsh`).

### Brewfiles

- `ansible/roles/workstations/files/homebrew/.Brewfile-mac` — macOS packages
- `ansible/roles/workstations/files/homebrew/.Brewfile-linux` — LinuxBrew packages

Ansible copies the appropriate Brewfile to `~/.Brewfile` based on detected OS, then runs `brew bundle install --global`.
