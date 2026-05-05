# CLAUDE.md

## What This Is

A personal private system bootstrap and configuration automation project using a symlink-based dotfiles approach. Manages developer machine setup (macOS MBP 2022, Linux LEMP9 server) by symlinking configs from this repo into `~/`.

Commit directly to `main` — no feature branches or PRs unless explicitly requested.

## Commands

**Install dotfiles (symlink into ~/):**
```shell
./install.sh
# or
make install
```

**Brewfile management:**
```shell
make brew-install   # install packages from dotfiles/.Brewfile
make brew-sync      # show drift between live brew state and dotfiles/.Brewfile
make brew-dump      # write live brew state back to dotfiles/.Brewfile (~/.Brewfile)
```

**Quick Zsh-only copy (no symlinks):**
```shell
cp -r ./dotfiles/.zsh* ~/
```

## Architecture

### Directory Structure

- `dotfiles/` — shared config files mirroring `~/`; `install.sh` symlinks each file into place
- `hosts/<machine>/` — machine-specific overrides; applied after shared dotfiles (detected via `hostname -s`)
- `install.sh` — idempotent symlink installer; backs up conflicts as `<file>.bak.<timestamp>`
- `Makefile` — convenience targets for install and Brewfile management

### Zsh Configuration

Located in `dotfiles/`:
- `.zshrc` — entry point, sources files from `.zsh/`
- `.zsh/` — modular configs loaded in numeric order (`0_path.zsh`, `0000_before.zsh`, etc.)
- `.zsh.before/` / `.zsh.after/` — hook directories for pre/post configs

Each tool gets its own file (e.g., `pyenv.zsh`, `nvm.zsh`, `golang.zsh`, `aliases.zsh`).

### Brewfile

- `dotfiles/.Brewfile` — symlinked to `~/.Brewfile`; used by `brew bundle install --global`

Use `make brew-dump` to capture live state back into the repo file.


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
