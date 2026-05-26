# bs CLI Design

**Date:** 2026-05-24
**Issue:** bs-cz1
**Status:** Approved

## Goal

Replace the ad-hoc mix of Make targets and shell scripts with a single, discoverable CLI binary (`bs`) for ongoing operational interaction with the dotfiles system. Machine setup scripts (`install.sh`, `macos-defaults.sh`, `init-tmux.sh`) are explicitly out of scope — they remain in bash/make where they belong.

## Language + Framework

**Go + Cobra.** Reasons: single self-contained binary after build, flag parsing and `--help` generation at every level for free, table-driven unit tests without a test framework, no runtime dependencies post-compile.

Go is available via mise (see mise migration issue).

## Project Layout

```
cli/                             # top-level directory, own go.mod
├── go.mod                       # module: github.com/dlstadther/bootstrap/cli
├── main.go                      # entry point, wires root command
├── cmd/
│   ├── root.go                  # root Cobra command, global flags, top-level help
│   ├── version.go               # bs version
│   ├── audit.go                 # bs audit [--all]
│   ├── brew/
│   │   ├── brew.go              # bs brew (group — prints subcommand list when bare)
│   │   ├── sync.go              # bs brew sync
│   │   ├── dump.go              # bs brew dump
│   │   └── install.go           # bs brew install
│   └── tmux/
│       ├── tmux.go              # bs tmux (group — prints subcommand list when bare)
│       └── workspace.go         # bs tmux add [--name --cwd --dev --agent]
└── internal/
    ├── brew/                    # Sync(), Dump(), Install() logic + tests
    ├── tmux/                    # workspace generation logic + tests
    ├── audit/                   # audit logic + tests
    ├── git/                     # rev-parse, dirty check, repo path resolution + tests
    └── version/                 # build-time constants injected via ldflags
```

`cmd/` is Cobra plumbing only: flag definitions, arg validation, delegation to `internal/`. All logic and all tests live in `internal/`. Nothing in `internal/` imports `cmd/`.

## Naming + Renameability

The binary name appears in exactly one place: the Makefile build target output flag (`-o ~/.local/bin/bs`). Cobra derives the name shown in help text from `os.Args[0]` automatically. Renaming `bs` to anything else is a one-line Makefile change — no internal packages reference the binary name.

## Version Command

`bs version` shows two values:

```
compiled: abc1234  (2026-05-24T12:00:00Z)
repo:     abc1234  (clean)
```

When the repo has moved forward since the last `make install`:

```
compiled: abc1234  (2026-05-24T12:00:00Z)
repo:     def5678  (dirty — run 'make install' to update)
```

**Build-time values** are injected via ldflags at `make install` into `internal/version`:
- `CommitHash` — `git rev-parse HEAD`
- `BuildTime` — UTC timestamp
- `RepoPath` — absolute path to the repo at build time (e.g. `$HOME/code/bootstrap`)

**Runtime repo path resolution** (`internal/git.RepoPath()`):
1. `$BOOTSTRAP_REPO` env var if set
2. Compiled-in `RepoPath` as fallback

This makes it work with zero config on machines that use the standard path, with an escape hatch for non-standard clones.

## Makefile Integration

```makefile
BS_LDFLAGS = -X cli/internal/version.CommitHash=$(shell git rev-parse HEAD) \
             -X cli/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
             -X cli/internal/version.RepoPath=$(HOME)/code/bootstrap

build-bs:
	cd cli && go build -ldflags "$(BS_LDFLAGS)" -o ~/.local/bin/bs .

test-bs:
	cd cli && go test ./...

install: build-bs
	./scripts/install.sh
```

`install` gains `build-bs` as a prerequisite — one `make install` sets up both the symlinked dotfiles and the binary.

`~/.local/bin` is already exported via `dotfiles/.zsh/0_path.zsh`, so the binary is on `$PATH` immediately after install with no further changes needed.

## Testing

`internal/` packages expose plain Go functions. Tests call them directly — no Cobra involved, no CLI subprocess overhead.

Commands that shell out (`brew`, `git`) accept an injectable `Executor` interface:

```go
type Executor interface {
    Run(cmd string, args ...string) (string, error)
}
```

Production code passes a real executor; tests pass a fake. This keeps subprocess calls out of unit tests entirely without requiring filesystem mocking.

## Initial Subcommand Scope

| Command | Description |
|---|---|
| `bs version` | Show compiled hash + current repo hash, flag drift |
| `bs audit [--all]` | Audit gap between repo and machine state |
| `bs brew sync` | Show drift between live brew state and .Brewfile |
| `bs brew dump` | Write live brew state back to .Brewfile |
| `bs brew install` | Install packages from .Brewfile |
| `bs tmux add` | Create a tmux workspace session (real logic, see below) |

### `bs tmux add` flags (initial)

| Flag | Description |
|---|---|
| `--name` | Session name |
| `--cwd` | Working directory for the session |
| `--dev` | Boolean — open a dev window layout |
| `--agent` | Agent to launch (e.g. `claude`) |

Logic lives in `internal/tmux/`. Design of the workspace layout is detailed when this subcommand is implemented.

## Out of Scope

- `install.sh` — dotfile symlinking, stays in bash/make
- `macos-defaults.sh` — macOS system config, stays in bash/make
- `init-tmux.sh` — tmux plugin setup, stays in bash/make

These are machine setup operations that run at the same time as the binary build/install. There is no benefit to wrapping them in the CLI.

## Cleanup (on completion)

Once all initial subcommands are implemented and verified, delete the scripts they supersede:

| File | Superseded by |
|---|---|
| `scripts/audit.sh` | `bs audit` |
| `dotfiles/.config/tmux/templates/workspace.sh` | `bs tmux add` |

The `make audit` target (if present) should also be removed from the Makefile at that point.
