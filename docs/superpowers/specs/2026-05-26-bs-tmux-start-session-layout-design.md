# Design: `bs tmux start` — Session Skip + Layout Clarity

**Date:** 2026-05-26  
**Status:** Approved

## Problem

Two bugs in `bs tmux start`:

1. **Existing session handling:** When tmux-resurrect restores a session before `bs tmux start` applies templates, `createPanesFromConfig` runs unconditionally on already-populated sessions. This creates duplicate panes (original resurrected + newly split), double-sends commands, and produces an unpredictable layout.

2. **YAML layout ambiguity:** The session YAML schema has no way to express pane size. `main-vertical` layout relies on tmux's default `main-pane-width` (80 cells), which doesn't reflect the intended 60/40 split. Additionally, `bootstrap.yaml` had the main pane percent coded incorrectly (40% instead of the intended 60%).

## Design

### Fix 1: Skip existing sessions in `createSessionFromConfig`

Add an early return at the top of `createSessionFromConfig`:

```go
func createSessionFromConfig(s SessionConfig, exec Executor) error {
    if sessionExists(s.Name, exec) {
        return nil  // leave resurrected sessions untouched
    }
    // ... rest of creation logic
}
```

**Behavior matrix:**

| Session state | `--override` | Result |
|---|---|---|
| Does not exist | any | Create from template (current behavior) |
| Already exists | false (default) | Skip entirely — leave as-is |
| Already exists | true | Kill then recreate from template (current behavior) |

### Fix 2: Add `main_pane_percent` to `WindowConfig`

New optional YAML field `main_pane_percent` (int, 0–100). Parsed into `WindowConfig.MainPanePercent`.

After all panes are created in `createPanesFromConfig`, if `MainPanePercent > 0` and a layout is set, apply:

```go
exec.Run("tmux", "set-window-option", "-t", target, "main-pane-width",
    fmt.Sprintf("%d%%", w.MainPanePercent))
exec.Run("tmux", "select-layout", "-t", target, w.Layout)
```

Requires tmux ≥ 3.1 (percentage syntax for `main-pane-width`).

### Fix 3: Correct `bootstrap.yaml`

```yaml
name: bootstrap
root: ~/code/bootstrap
windows:
  - main:
      layout: main-vertical       # pane 0 = large LEFT, panes 1+ stack vertically on RIGHT
      main_pane_percent: 60       # left (claude) gets 60%, right gets 40%
      panes:
        - command: "claude agents --cwd ~/code/bootstrap"
          no_enter: true          # pane 0 → left (60%)
        - git pull origin main && ls -al && bd ready  # pane 1 → right top
        - lazygit                 # pane 2 → right bottom
```

### Fix 4: Align test fixtures

Update `basicYAML` in `session_test.go` to match actual `bootstrap.yaml` pane order (claude first with `no_enter: true`, git pull second, lazygit third) and add `main_pane_percent: 60`. Update test assertions to match.

Add two new tests:
- `TestParseSession_MainPanePercent` — `main_pane_percent: 60` parses to `MainPanePercent: 60`
- `TestStart_SkipsExistingSession` — when `has-session` succeeds, no `split-window` or `send-keys` calls are made

## Files Changed

| File | Change |
|---|---|
| `cli/internal/tmux/session.go` | Add `MainPanePercent` to `WindowConfig`; early return in `createSessionFromConfig`; apply `main-pane-width` after pane creation |
| `cli/internal/tmux/session_test.go` | Update `basicYAML`, update assertions, add 2 new tests |
| `hosts/mbp2022/.config/tmux/sessions/bootstrap.yaml` | Add `main_pane_percent: 60`, add layout comments |

## Out of Scope

- `workspace.go` (`bs tmux add` flow) — separate from session YAML
- `admin.yaml` — no layout, no change needed
- `--override` behavior — already correct
- Explicit split-tree schema (deferred as future enhancement)
