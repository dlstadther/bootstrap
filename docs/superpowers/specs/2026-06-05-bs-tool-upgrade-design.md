# Design: `bs tool upgrade`

**Date:** 2026-06-05
**Status:** Approved (pending spec review)

## Problem

The dotfiles assume a handful of **top-level CLI tools** are already installed on the
machine. These are not managed by the repo's package flows — they are not in any
Brewfile, not in the repo's `.mise.toml`, and not in the dotfiles' global mise config
(`~/.config/mise/config.toml`). Today there is no single command to check whether these
tools are current and upgrade them.

We want a command that:

1. Checks the installed version of each top-level tool against its latest available version.
2. Prompts the user yes/no for **every** tool **before** applying any upgrade, so the user
   can answer all prompts and then walk away while upgrades run unattended.
3. Applies only the approved upgrades.
4. Offers a non-interactive check-only mode for a quick status glance.

The user explicitly does **not** want bleeding-edge forced upgrades: every tool is an
individual opt-in each run.

## Scope

### In scope (top-level, unmanaged tools)

| Tool      | Current-version source | Latest-version source                          | Upgrade action     |
|-----------|------------------------|------------------------------------------------|--------------------|
| `brew`    | `brew --version`       | GitHub `Homebrew/brew` latest release tag      | `brew update`      |
| `claude`  | `claude --version`     | unknown (no clean pre-check) → determined on apply | `claude update` |
| `opencode`| `opencode --version`   | best-effort (see Open Questions)               | `opencode upgrade` |

`brew update` updates Homebrew itself and its formula/tap metadata — **not** installed
packages (that is "content", explicitly out of scope).

### Explicitly out of scope

- **mise, gh, gemini-cli, git, tmux, vim, prek, cmux, codex** — installed via the Brewfile;
  upgraded through the existing `bs brew` flow. (`mise self-update` also refuses to run on a
  brew install, so `brew upgrade mise` is the correct path anyway.)
- **uv, rustup, node, go, bun, pnpm, rust, python, etc.** — managed by mise.
- **bs** — built from this repo via `make install`; not a third-party tool.
- **Installed-package upgrades** (`brew upgrade`, `mise up`) — that is managed content.
- **rbenv** — referenced in `rbenv.zsh` but guarded by `command -v`, not installed, PATH
  line commented out. Dormant; not a target.

### Related cleanup (filed separately, NOT part of this feature)

- **bs-4sj** — remove vestigial pyenv install + leftover `$HOME/.pyenv/bin` PATH entry.
- **bs-b61** — reconcile gcloud SDK path mismatch (manual install vs brew cask).

## Approach

A **registry-based Go package**, consistent with the existing `internal/audit` and
`internal/brew` packages (mockable `Executor` interface, unit tests alongside). Each tool
is one registry entry implementing a small interface; a `Run` orchestrator handles
check → decide → apply. Adding a future tool is a single registry entry.

Alternatives considered and rejected:
- **Hardcoded inline logic** — every new tool edits the core `Run` function; messier tests.
- **Pure shell script** (`scripts/upgrade.sh` + make target) — no unit tests, diverges from
  the `bs` CLI, and the "collect all answers then apply" split plus version parsing are
  clumsier in bash.

## Command Surface

- `bs tool` — new group command; prints help when invoked bare (like `bs brew`).
- `bs tool upgrade` — interactive flow (check → prompt-all → apply → summary).
- `bs tool upgrade --check` — print the status table and exit; no prompts, no upgrades.

## Architecture

### Packages (mirrors existing layout)

- `cli/cmd/tool/tool.go` — `bs tool` group.
- `cli/cmd/tool/upgrade.go` — `bs tool upgrade` command + `--check` flag; wires to internal.
- `cli/internal/toolupgrade/` — core logic:
  - `Executor` interface (same shape as `internal/brew` / `internal/audit`).
  - `Tool` interface (below) + concrete `brew`, `claude`, `opencode` implementations.
  - `registry` — ordered slice of `Tool`.
  - `Run(opts, exec, decider)` orchestrator.

### `Tool` interface

```go
type Tool interface {
    Name() string
    Installed(exec Executor) bool              // false → warn + skip (expected but missing)
    CurrentVersion(exec Executor) (string, error)
    LatestVersion(exec Executor) (string, error) // "" + nil ⇒ unknown (not an error)
    Upgrade(exec Executor) error               // idempotent: no-op if already current
}
```

### Status row

```go
type Status struct {
    Name    string
    Current string
    Latest  string // "" rendered as "—"
    State   State  // UpToDate | UpdateAvailable | Unknown | NotInstalled
}
```

`State` derivation:
- both versions known and equal → `UpToDate`
- both known and different → `UpdateAvailable`
- latest unknown (`""`) but installed → `Unknown`
- not installed → `NotInstalled` (warn, never prompted)

## Data Flow (enforces "ask everything first")

1. For each registered tool: if `Installed`, gather current + latest into a `Status` row;
   otherwise mark `NotInstalled`.
2. Print the status table: `TOOL | CURRENT | LATEST | STATUS`.
3. If `--check`: stop here (exit 0).
4. Build the candidate list = tools in state `UpdateAvailable` or `Unknown`.
5. **`Decider` collects ALL yes/no answers up front** (one prompt per candidate) and returns
   the approved set. No upgrade runs during this phase — this structurally guarantees the
   user finishes all prompts before any work begins.
6. Apply approved upgrades **sequentially**, printing per-tool progress.
7. Print summary: `N upgraded, N skipped, N failed`.

### `Decider` seam (for testability)

```go
type Decider func(candidates []Status) (approved map[string]bool, err error)
```

- Production default: reads `y/N` from stdin for each candidate (defaults to No on empty/EOF).
- Tests inject a `Decider` that returns a fixed approval map — no stdin needed.

## Example

```
$ bs tool upgrade
Checking top-level tools…

  TOOL       CURRENT    LATEST     STATUS
  brew       4.6.0      4.6.2      update available
  claude     2.1.165    —          unknown (checks on upgrade)
  opencode   1.2.10     1.2.13     update available

Upgrade brew (4.6.0 → 4.6.2)? [y/N] y
Upgrade claude (2.1.165 → ?)? [y/N] n
Upgrade opencode (1.2.10 → 1.2.13)? [y/N] y

Applying 2 upgrades…
  → brew … done
  → opencode … done (1.2.13)

Summary: 2 upgraded, 1 skipped, 0 failed.
```

```
$ bs tool upgrade --check
  TOOL       CURRENT    LATEST     STATUS
  brew       4.6.0      4.6.2      update available
  claude     2.1.165    —          unknown
  opencode   1.2.10     1.2.13     update available
```

## Error Handling

- **Tool not installed:** print a warning line, mark `NotInstalled`, never prompt. Does not
  fail the run.
- **Latest-version fetch fails** (e.g. no network): treat as `Unknown` (latest = `""`), not a
  fatal error. The tool is still offered for opt-in upgrade because the upgrade commands are
  idempotent.
- **An upgrade command fails:** record it as `failed` in the summary with the error, and
  **continue** to the next approved tool — one failure must not abort the batch.
- **Non-interactive stdin without `--check`:** the default `Decider` reads No on EOF, so no
  upgrades run; the table still prints. (A `--yes` flag is explicitly out of v1.)

## Testing

`cli/internal/toolupgrade/*_test.go` with a mock `Executor` and an injected `Decider`:

- Version parsing per tool (`brew --version` → `4.6.0`; `claude --version`, `opencode --version`).
- `State` derivation: equal → UpToDate; differing → UpdateAvailable; empty latest → Unknown.
- **Only approved tools upgrade** — Decider approves a subset; assert exactly those
  `Upgrade` commands ran (and no upgrade ran before any prompt).
- Failure handling — one tool's `Upgrade` errors; assert summary counts and that remaining
  approved tools still ran.
- `--check` mode runs no upgrade commands.
- `NotInstalled` tools are skipped, not prompted.

## YAGNI / Out of v1

- No `--yes` / non-interactive auto-approve flag.
- No persisted skip/pin list (every run prompts fresh — confirmed preference).
- No `make` wrapper target (`bs tool upgrade` is the entry point).

## Open Questions (resolve during implementation, not from memory)

- **claude:** confirm the exact current-version command (`claude --version` output format) and
  that `claude update` is the right upgrade verb. No clean pre-check for latest is assumed →
  shown as `—`.
- **opencode:** confirm `opencode --version` output format and whether a clean "latest"
  source exists (e.g. a version manifest or GitHub releases for `sst/opencode`). If not,
  fall back to `—` like claude; `opencode upgrade` still works idempotently.
- **brew latest:** fetch `Homebrew/brew` latest release tag via `curl` through the `Executor`
  (so it stays mockable) and parse `tag_name`; tolerate failure → `Unknown`.
