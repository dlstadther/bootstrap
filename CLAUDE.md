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
bs brew diff          # show drift between live brew state and .Brewfile
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

Homebrew 6+ refuses to load formulae from untrusted third-party taps. Trust is stored in `~/.homebrew/trust.json`; track it per host (e.g. `hosts/<machine>/.homebrew/trust.json`) so a fresh bootstrap trusts the declared taps. Only **formula** taps need trusting — cask taps do not. Add a tap with `brew trust <tap>`.

### Tool Installation

Prefer **mise** over Homebrew for versioned tools (runtimes, CLIs, language toolchains). Mise makes version pinning, minimum version age constraints, and per-project overrides easier than brew. Use Homebrew for GUI apps, system-level tools, and anything without a stable mise backend.

- `dotfiles/.config/mise/config.toml` — symlinked to `~/.config/mise/config.toml`; global mise tool versions

### Claude Plugins

Claude Code plugins are not tracked in dotfiles directly — they live in `~/.claude/plugins/` (managed by Claude Code). To reinstall on a fresh machine, run `make install-plugins` after `make install`.

**How it works:**
- `make install-plugins` runs `scripts/install-plugins.sh`
- The script reads `enabledPlugins` from `~/.claude/settings.json` via `jq` and installs each enabled plugin
- Non-official marketplace sources are declared in `extraKnownMarketplaces` in the appropriate `settings.json`:
  - `dotfiles/.claude/settings.json` — shared across all hosts
  - `hosts/<machine>/.claude/settings.json` — host-specific additions

**`extraKnownMarketplaces` format:**
```json
"extraKnownMarketplaces": {
  "marketplace-id": {
    "source": { "source": "github", "repo": "owner/repo" }
  }
}
```

When adding a new non-official marketplace, add it to `extraKnownMarketplaces` in the appropriate `settings.json` and enable its plugins in `enabledPlugins`.

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

## CMUX Reference Material

When working with CMUX (commands, API, configuration, or behavior), **always consult the latest published web documentation** at https://cmux.com/docs/api before answering or making changes.

- **Do NOT rely on training data or memory** for CMUX details. The API and CLI change, and answers reconstructed from memory have repeatedly been wrong.
- **Fetch the live docs every time.** Use the `find-docs` skill or fetch https://cmux.com/docs/api directly to confirm command names, flags, surface/pane semantics, and request/response shapes before acting.
- Verify against the docs even when something "looks obvious" — confirm the exact syntax rather than guessing.


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

### Cross-Machine Sync

**On a new machine (fresh clone):** use `bd bootstrap` — NOT `bd init`. Bootstrap clones the dolt DB from the remote. `bd init` creates an independent local history that can never merge with the remote ("no common ancestor").

**Each session:**
```bash
bd dolt pull          # pull at session start (if working across machines)
# ... do work ...
bd dolt push          # push at session end (before git push)
git push
```

**If dolt remote is missing** (symptom: `bd dolt pull` fails with "no remote"):
```bash
bd dolt remote add origin "git+ssh://git@github.com/dlstadther/bootstrap.git"
bd dolt pull
```

**If histories diverged** (symptom: "no common ancestor"):
```bash
rm -rf .beads/embeddeddolt/bootstrap
bd bootstrap --yes    # re-clone from remote; local-only closed issues are lost
```

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
