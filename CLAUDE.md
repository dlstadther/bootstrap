# CLAUDE.md

## What This Is

A personal private system bootstrap and configuration automation project using Ansible. Automates developer machine setup (macOS MBP 2022, Linux LEMP9 server) including Zsh configuration, Homebrew packages, Vim, and JetBrains Toolbox.

Commit directly to `main` — no feature branches or PRs unless explicitly requested.

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


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
