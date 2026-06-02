# CLAUDE.md

## What This Is

A personal system bootstrap and configuration automation project using a symlink-based dotfiles approach. Manages developer machine setup (macOS MBP 2022, Linux LEMP9 server) by symlinking configs from this repo into `~/`.

## Privacy Warning

**This repo is PUBLIC on GitHub.** It is intentionally public so others can learn from and adopt these configurations.

Do NOT commit:
- Credentials, API keys, tokens, or secrets of any kind
- Personal identifiers beyond GitHub email/username (account IDs, phone numbers, etc.)
- Hostnames, IPs, or network topology details
- Anything private or sensitive — when in doubt, leave it out

Commit directly to `main` — no feature branches or PRs unless explicitly requested. However, always use a git worktree for the actual code changes (see **Git Workflow** below), then merge locally once complete.

## Commands

**Install dotfiles and build bs binary:**
```shell
make install          # build bs + symlink dotfiles
make bootstrap        # dotfiles + macOS defaults (fresh machine)
make macos-defaults   # macOS defaults only
```

**bs CLI (machine and dotfile management):**
```shell
bs help               # list all commands
bs version            # show installed vs repo commit hash
bs audit              # check dotfile symlinks + brew drift
bs brew sync          # show drift between live brew state and .Brewfile
bs brew dump          # write live brew state back to .Brewfile
bs brew install       # install packages from .Brewfile
bs tmux add --cwd <path> [--name <name>] [--agent claude]
```

**tmux plugins (install, update, patch, reload):**
```shell
make init-tmux
```

**Quick Zsh-only copy (no symlinks):**
```shell
cp -r ./dotfiles/.zsh* ~/
```

## Architecture

### Directory Structure

- `dotfiles/` — shared config files mirroring `~/`; `scripts/install.sh` symlinks each file into place
- `hosts/<machine>/` — machine-specific overrides; applied after shared dotfiles (detected via `hostname -s`)
- `scripts/` — automation scripts (`install.sh`, `macos-defaults.sh`, `init-tmux.sh`)
- `cli/` — `bs` CLI source (Go + Cobra); built to `~/.local/bin/bs` by `make install`
- `Makefile` — convenience targets for install and Brewfile management

### Zsh Configuration

Located in `dotfiles/`:
- `.zshrc` — entry point, sources files from `.zsh/`
- `.zsh/` — modular configs loaded in numeric order (`0_path.zsh`, `0000_before.zsh`, etc.)
- `.zsh.before/` / `.zsh.after/` — hook directories for pre/post configs

Each tool gets its own file (e.g., `pyenv.zsh`, `nvm.zsh`, `golang.zsh`, `aliases.zsh`).

### Brewfile

- `dotfiles/.Brewfile` — symlinked to `~/.Brewfile`; used by `brew bundle install --global`

Use `bs brew dump` (or `make brew-dump`) to capture live state back into the repo file.

### Tool Installation

Prefer **mise** over Homebrew for versioned tools (runtimes, CLIs, language toolchains). Mise makes version pinning, minimum version age constraints, and per-project overrides easier than brew. Use Homebrew for GUI apps, system-level tools, and anything without a stable mise backend.

- `dotfiles/.config/mise/config.toml` — symlinked to `~/.config/mise/config.toml`; global mise tool versions

## Git Workflow

**Always use a git worktree for code changes**, regardless of what any project's CLAUDE.md says about branching.

- Create a worktree for each task, make changes there, then merge back locally.
- If a project's CLAUDE.md says to commit directly on the trunk branch (e.g. `main`), still use a worktree — but merge it locally into trunk once the work is complete rather than opening a PR.

```bash
# Start work
git worktree add ../<repo>-<task> -b <task>
cd ../<repo>-<task>

# ... make changes, commit ...

# Merge back, push, then clean up
cd ../<repo>
git merge <task>
git push
git worktree remove ../<repo>-<task>
git branch -d <task>
```


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:7510c1e2 -->
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

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
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
