# bs tmux start — Session Skip + Layout Clarity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix two bugs in `bs tmux start`: (1) stop creating duplicate panes when a session already exists from tmux-resurrect, and (2) add a `main_pane_percent` YAML field that sets the actual tmux main pane width.

**Architecture:** All changes are in `cli/internal/tmux/session.go` (Go struct + parsing + execution logic) and its test file. `bootstrap.yaml` is updated separately to use the new field with the correct 60% value.

**Tech Stack:** Go 1.24, `gopkg.in/yaml.v3`, tmux ≥ 3.1 (for `main-pane-width` percentage syntax), Cobra CLI.

---

## File Map

| File | What changes |
|---|---|
| `cli/internal/tmux/session.go` | Add `MainPanePercent int` to `WindowConfig`; parse `main_pane_percent` YAML key; add early return to `createSessionFromConfig`; call `set-window-option main-pane-width` at end of `createPanesFromConfig` |
| `cli/internal/tmux/session_test.go` | Update `basicYAML` fixture to match real `bootstrap.yaml` pane order + add `main_pane_percent: 60`; fix two existing assertions; add `TestParseSession_MainPanePercent` and `TestStart_SkipsExistingSession` and `TestStart_AppliesMainPaneWidth` |
| `hosts/mbp2022/.config/tmux/sessions/bootstrap.yaml` | Add `main_pane_percent: 60`; add layout comments clarifying pane positions |

---

## Task 1: Update `basicYAML` fixture and fix dependent test assertions

The test fixture `basicYAML` in `session_test.go` has a different pane order than the real `bootstrap.yaml` (claude is at index 2 in the test, index 0 in the real file). Align them first, so later tasks build on an accurate fixture.

**Files:**
- Modify: `cli/internal/tmux/session_test.go`

- [ ] **Step 1: Replace `basicYAML` with the corrected fixture**

Open `cli/internal/tmux/session_test.go`. Replace the `basicYAML` const (lines ~14–25) with:

```go
const basicYAML = `
name: bootstrap
root: ~/code/bootstrap
windows:
  - main:
      layout: main-vertical
      main_pane_percent: 60
      panes:
        - command: "claude agents --cwd ~/code/bootstrap"
          no_enter: true
        - git pull origin main && ls -al && bd ready
        - lazygit
`
```

- [ ] **Step 2: Fix `TestParseSession_PaneCommands`**

The test currently checks `panes[0].Command == "git pull origin main && bd ready"` and `panes[0].NoEnter == false`. With the new fixture, pane[0] is claude. Replace the assertions:

```go
func TestParseSession_PaneCommands(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	panes := s.Windows[0].Panes
	if len(panes) != 3 {
		t.Fatalf("panes: want 3, got %d", len(panes))
	}
	if panes[0].Command != "claude agents --cwd ~/code/bootstrap" {
		t.Errorf("pane[0]: want claude command, got %q", panes[0].Command)
	}
	if !panes[0].NoEnter {
		t.Error("pane[0]: no_enter should be true")
	}
}
```

- [ ] **Step 3: Fix `TestParseSession_PaneNoEnter`**

The test currently reads `pane := s.Windows[0].Panes[2]` (claude was at index 2). Update to index 0:

```go
func TestParseSession_PaneNoEnter(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	pane := s.Windows[0].Panes[0]
	if pane.Command != "claude agents --cwd ~/code/bootstrap" {
		t.Errorf("pane[0] command: got %q", pane.Command)
	}
	if !pane.NoEnter {
		t.Error("pane[0]: no_enter should be true")
	}
}
```

- [ ] **Step 4: Run existing tests — expect compile failure**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -v 2>&1 | head -40
```

Expected: compile error referencing `MainPanePercent` (unknown field) — that's fine, we add it in Task 2.

---

## Task 2: Add `MainPanePercent` to `WindowConfig` and parse it from YAML (TDD)

**Files:**
- Modify: `cli/internal/tmux/session_test.go`
- Modify: `cli/internal/tmux/session.go`

- [ ] **Step 1: Write the failing test**

Add this test to `cli/internal/tmux/session_test.go` after `TestParseSession_PaneNoEnter`:

```go
func TestParseSession_MainPanePercent(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if s.Windows[0].MainPanePercent != 60 {
		t.Errorf("main_pane_percent: want 60, got %d", s.Windows[0].MainPanePercent)
	}
}
```

- [ ] **Step 2: Run to confirm compile failure (field not yet defined)**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestParseSession_MainPanePercent -v 2>&1
```

Expected: `undefined: tmux.WindowConfig.MainPanePercent` or similar compile error.

- [ ] **Step 3: Add `MainPanePercent int` to `WindowConfig` in `session.go`**

In `cli/internal/tmux/session.go`, update `WindowConfig`:

```go
// WindowConfig represents a single tmux window within a session.
type WindowConfig struct {
	Name            string
	Root            string
	Layout          string
	MainPanePercent int // if > 0, set main-pane-width to this % after applying layout (tmux >= 3.1)
	Panes           []PaneConfig
}
```

- [ ] **Step 4: Parse `main_pane_percent` in `parseWindowNode`**

In `session.go`, inside `parseWindowNode`, find the `case yaml.MappingNode:` block. The `detail` struct currently has `Root`, `Layout`, `Panes`. Add `MainPanePercent`:

```go
case yaml.MappingNode:
	var detail struct {
		Root            string      `yaml:"root"`
		Layout          string      `yaml:"layout"`
		MainPanePercent int         `yaml:"main_pane_percent"`
		Panes           []yaml.Node `yaml:"panes"`
	}
	if err := valueNode.Decode(&detail); err != nil {
		return WindowConfig{}, err
	}
	wc.Root = detail.Root
	wc.Layout = detail.Layout
	wc.MainPanePercent = detail.MainPanePercent
	for i := range detail.Panes {
		pc, err := parsePaneNode(&detail.Panes[i])
		if err != nil {
			return WindowConfig{}, err
		}
		wc.Panes = append(wc.Panes, pc)
	}
```

- [ ] **Step 5: Run the test — expect PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestParseSession_MainPanePercent -v
```

Expected:
```
--- PASS: TestParseSession_MainPanePercent (0.00s)
PASS
```

- [ ] **Step 6: Run all tests — expect all PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -v
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
cd /Users/dlstadther/code/bootstrap
git add cli/internal/tmux/session.go cli/internal/tmux/session_test.go
git commit -m "feat(bs-tmux): add main_pane_percent field to WindowConfig YAML schema"
```

---

## Task 3: Skip existing sessions in `createSessionFromConfig` (TDD)

**Files:**
- Modify: `cli/internal/tmux/session_test.go`
- Modify: `cli/internal/tmux/session.go`

- [ ] **Step 1: Write the failing test**

Add this test to `cli/internal/tmux/session_test.go` after the `TestStart_*` group:

```go
func TestStart_SkipsExistingSession(t *testing.T) {
	exec := newFake()
	// Session "bootstrap" exists — has-session succeeds (no error, returns "")
	exec.results["tmux has-session"] = ""

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bootstrap.yaml"), []byte(`
name: bootstrap
root: ~/code/bootstrap
windows:
  - main:
      layout: main-vertical
      panes:
        - command: "claude agents"
          no_enter: true
        - git status
`), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 {
			if c.args[0] == "split-window" {
				t.Errorf("should not call split-window when session already exists; got call: %v", c.args)
			}
			if c.args[0] == "send-keys" {
				t.Errorf("should not call send-keys when session already exists; got call: %v", c.args)
			}
		}
	}
}
```

- [ ] **Step 2: Run to confirm FAIL**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestStart_SkipsExistingSession -v
```

Expected: `FAIL` — `should not call send-keys when session already exists`.

- [ ] **Step 3: Add early return to `createSessionFromConfig` in `session.go`**

At the top of `createSessionFromConfig`, before the `for i, w := range s.Windows` loop:

```go
func createSessionFromConfig(s SessionConfig, exec Executor) error {
	// If the session already exists, leave it untouched.
	// tmux-resurrect may have restored its layout — adding panes would create duplicates.
	if sessionExists(s.Name, exec) {
		return nil
	}

	for i, w := range s.Windows {
		// ... rest of existing code unchanged
```

- [ ] **Step 4: Run the new test — expect PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestStart_SkipsExistingSession -v
```

Expected:
```
--- PASS: TestStart_SkipsExistingSession (0.00s)
PASS
```

- [ ] **Step 5: Run all tests — expect all PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -v
```

Expected: all tests pass. Note: `TestStart_Override_KillsMatchingSessions` sets `exec.results["tmux has-session"] = ""` AND `Override: true`, so the session gets killed before `createSessionFromConfig` is called — the early return won't interfere.

- [ ] **Step 6: Commit**

```bash
cd /Users/dlstadther/code/bootstrap
git add cli/internal/tmux/session.go cli/internal/tmux/session_test.go
git commit -m "fix(bs-tmux): skip pane creation when session already exists (prevents resurrect duplicates)"
```

---

## Task 4: Apply `main-pane-width` after pane creation in `createPanesFromConfig` (TDD)

**Files:**
- Modify: `cli/internal/tmux/session_test.go`
- Modify: `cli/internal/tmux/session.go`

- [ ] **Step 1: Write the failing test**

Add this test to `cli/internal/tmux/session_test.go`:

```go
func TestStart_AppliesMainPaneWidth(t *testing.T) {
	exec := newFake()
	// no existing sessions
	exec.errs["tmux has-session"] = errFake

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "work.yaml"), []byte(`
name: work
root: ~/code/work
windows:
  - main:
      layout: main-vertical
      main_pane_percent: 60
      panes:
        - git status
        - lazygit
`), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range exec.calls {
		if c.cmd != "tmux" || len(c.args) < 2 || c.args[0] != "set-window-option" {
			continue
		}
		for i, arg := range c.args {
			if arg == "main-pane-width" && i+1 < len(c.args) && c.args[i+1] == "60%" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected set-window-option main-pane-width 60% to be called")
	}
}
```

- [ ] **Step 2: Run to confirm FAIL**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestStart_AppliesMainPaneWidth -v
```

Expected: `FAIL` — `expected set-window-option main-pane-width 60% to be called`.

- [ ] **Step 3: Add `main-pane-width` application to `createPanesFromConfig` in `session.go`**

At the end of `createPanesFromConfig`, after the pane loop, add:

```go
	// Apply main pane size constraint after all panes are created.
	// Uses percentage syntax — requires tmux >= 3.1.
	if w.Layout != "" && w.MainPanePercent > 0 {
		exec.Run("tmux", "set-window-option", "-t", target, "main-pane-width", fmt.Sprintf("%d%%", w.MainPanePercent)) //nolint:errcheck
		exec.Run("tmux", "select-layout", "-t", target, w.Layout)                                                       //nolint:errcheck
	}

	return nil
}
```

The complete final state of `createPanesFromConfig` should be:

```go
func createPanesFromConfig(session string, w WindowConfig, defaultRoot string, exec Executor) error {
	target := session + ":" + w.Name
	root := coalesce(w.Root, defaultRoot)

	for i, p := range w.Panes {
		if i > 0 {
			if _, err := exec.Run("tmux", "split-window", "-t", target, "-c", expandHome(root)); err != nil {
				return fmt.Errorf("split-window: %w", err)
			}
			if w.Layout != "" {
				exec.Run("tmux", "select-layout", "-t", target, w.Layout) //nolint:errcheck
			}
		}
		// Target the active pane in the window — after split-window the new pane
		// is always active, so this correctly follows pane creation order without
		// hardcoding numeric indices (which break with pane-base-index != 0).
		if p.Command != "" {
			if p.NoEnter {
				exec.Run("tmux", "send-keys", "-t", target, p.Command) //nolint:errcheck
			} else {
				exec.Run("tmux", "send-keys", "-t", target, p.Command, "Enter") //nolint:errcheck
			}
		}
	}

	// Apply main pane size constraint after all panes are created.
	// Uses percentage syntax — requires tmux >= 3.1.
	if w.Layout != "" && w.MainPanePercent > 0 {
		exec.Run("tmux", "set-window-option", "-t", target, "main-pane-width", fmt.Sprintf("%d%%", w.MainPanePercent)) //nolint:errcheck
		exec.Run("tmux", "select-layout", "-t", target, w.Layout)                                                       //nolint:errcheck
	}

	return nil
}
```

- [ ] **Step 4: Run the new test — expect PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -run TestStart_AppliesMainPaneWidth -v
```

Expected:
```
--- PASS: TestStart_AppliesMainPaneWidth (0.00s)
PASS
```

- [ ] **Step 5: Run all tests — expect all PASS**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./cli/internal/tmux/... -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/dlstadther/code/bootstrap
git add cli/internal/tmux/session.go cli/internal/tmux/session_test.go
git commit -m "feat(bs-tmux): apply main-pane-width percentage after pane creation"
```

---

## Task 5: Update `bootstrap.yaml` with correct percent and layout comments

**Files:**
- Modify: `hosts/mbp2022/.config/tmux/sessions/bootstrap.yaml`

- [ ] **Step 1: Replace the file contents**

```yaml
name: bootstrap
root: ~/code/bootstrap
windows:
  - main:
      layout: main-vertical       # pane 0 = large LEFT; panes 1+ stack vertically on RIGHT
      main_pane_percent: 60       # left (claude) gets 60%, right stack gets 40%
      panes:
        - command: "claude agents --cwd ~/code/bootstrap"
          no_enter: true          # pane 0 → left 60% (staged; user launches)
        - git pull origin main && ls -al && bd ready  # pane 1 → right top
        - lazygit                 # pane 2 → right bottom
```

- [ ] **Step 2: Build the binary to confirm no compile errors**

```bash
cd /Users/dlstadther/code/bootstrap
make install 2>&1 | tail -5
```

Expected: build succeeds, no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/dlstadther/code/bootstrap
git add hosts/mbp2022/.config/tmux/sessions/bootstrap.yaml
git commit -m "fix(bs-tmux): set main_pane_percent to 60 and clarify pane layout in bootstrap.yaml"
```

---

## Task 6: Final verification

- [ ] **Step 1: Run the full test suite**

```bash
cd /Users/dlstadther/code/bootstrap
go test ./... -v 2>&1 | tail -20
```

Expected: all packages pass, no failures.

- [ ] **Step 2: Push**

```bash
cd /Users/dlstadther/code/bootstrap
git push
```

Expected: push succeeds, remote is up to date.
