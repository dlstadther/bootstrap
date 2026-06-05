# `bs tool upgrade` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `bs tool upgrade` command that checks the top-level CLI tools the dotfiles assume are installed (brew, claude, opencode), prompts the user yes/no for every out-of-date tool up front, then applies only the approved upgrades.

**Architecture:** A registry-based Go package `cli/internal/toolupgrade` (mockable `Executor` interface + `Tool` interface + `Run` orchestrator), wired through a new `cli/cmd/tool` Cobra group, mirroring the existing `internal/brew` + `cmd/brew` pattern. All prompts are collected by a `Decider` before any `Upgrade` runs, structurally guaranteeing "ask everything first."

**Tech Stack:** Go 1.26, Cobra, standard library (`text/tabwriter`, `bufio`, `encoding/json`). Module path: `github.com/dlstadther/bootstrap/cli`.

**Spec:** `docs/superpowers/specs/2026-06-05-bs-tool-upgrade-design.md`

---

## File Structure

**Create:**
- `cli/internal/toolupgrade/toolupgrade.go` — `Executor`, `State`, `Status`, `Tool`, `Options`, `Decider` types; `Evaluate`, `Run`, `renderTable`, `stateLabel`, `StdinDecider`.
- `cli/internal/toolupgrade/toolupgrade_test.go` — black-box tests for `Evaluate` + `Run` using fake `Tool`s and an injected `Decider`.
- `cli/internal/toolupgrade/tools.go` — concrete `brewTool`, `claudeTool`, `opencodeTool`; `Registry()`; helpers `firstField`, `firstLine`, `githubLatestTag`.
- `cli/internal/toolupgrade/tools_test.go` — black-box tests for each tool's version parsing + upgrade commands using a fake `Executor`.
- `cli/cmd/tool/tool.go` — `bs tool` group command + `realExecutor` (implements `Run` and `LookPath`).
- `cli/cmd/tool/upgrade.go` — `bs tool upgrade` command + `--check` flag.

**Modify:**
- `cli/cmd/root.go:21-26` — register `tool.Cmd`.

---

## Reference: real tool behavior (verified 2026-06-05)

| Tool | `--version` output | Latest source | Upgrade command |
|------|--------------------|---------------|-----------------|
| brew | `Homebrew 5.1.14` | `GET api.github.com/repos/Homebrew/brew/releases/latest` → `tag_name` `5.1.15` (no `v`) | `brew update` |
| claude | `2.1.165 (Claude Code)` | none → `""` (unknown) | `claude update` |
| opencode | `1.2.10` | `GET api.github.com/repos/sst/opencode/releases/latest` → `tag_name` `v1.16.0` (strip leading `v`) | `opencode upgrade` |

---

## Task 1: Core types + `Evaluate`

Defines the package vocabulary and the pure status-derivation function. No I/O, no concrete tools yet.

**Files:**
- Create: `cli/internal/toolupgrade/toolupgrade.go`
- Test: `cli/internal/toolupgrade/toolupgrade_test.go`

- [ ] **Step 1: Write the failing test**

Create `cli/internal/toolupgrade/toolupgrade_test.go`:

```go
package toolupgrade_test

import (
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)

// fakeTool is a controllable Tool for orchestration tests.
type fakeTool struct {
	name       string
	installed  bool
	current    string
	latest     string
	currentErr error
	latestErr  error
	upgradeErr error
	upgraded   bool
}

func (t *fakeTool) Name() string                                          { return t.name }
func (t *fakeTool) Installed(toolupgrade.Executor) bool                   { return t.installed }
func (t *fakeTool) CurrentVersion(toolupgrade.Executor) (string, error)   { return t.current, t.currentErr }
func (t *fakeTool) LatestVersion(toolupgrade.Executor) (string, error)    { return t.latest, t.latestErr }
func (t *fakeTool) Upgrade(toolupgrade.Executor) error {
	if t.upgradeErr != nil {
		return t.upgradeErr
	}
	t.upgraded = true
	return nil
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name string
		tool *fakeTool
		want toolupgrade.State
	}{
		{"not installed", &fakeTool{name: "x", installed: false}, toolupgrade.StateNotInstalled},
		{"up to date", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: "1.0.0"}, toolupgrade.StateUpToDate},
		{"update available", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: "1.1.0"}, toolupgrade.StateUpdateAvailable},
		{"latest unknown", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: ""}, toolupgrade.StateUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := toolupgrade.Evaluate(c.tool, nil)
			if got.State != c.want {
				t.Errorf("state: want %v, got %v", c.want, got.State)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/toolupgrade/ -run TestEvaluate -v`
Expected: FAIL — build error, `undefined: toolupgrade.Evaluate` / `toolupgrade.Executor` / `toolupgrade.State*`.

- [ ] **Step 3: Write minimal implementation**

Create `cli/internal/toolupgrade/toolupgrade.go`:

```go
// Package toolupgrade checks and upgrades top-level CLI tools the dotfiles
// assume are installed but do not manage via Brewfile or mise.
package toolupgrade

// Executor runs a command and returns combined output. LookPath reports whether
// a binary is resolvable on PATH. Both are seams for testing.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
	LookPath(name string) (string, error)
}

// State is the upgrade status of a tool.
type State int

const (
	StateUpToDate State = iota
	StateUpdateAvailable
	StateUnknown
	StateNotInstalled
)

// Status is the evaluated state of one tool.
type Status struct {
	Name    string
	Current string
	Latest  string // "" means unknown
	State   State
}

// Tool is one top-level managed binary.
type Tool interface {
	Name() string
	Installed(exec Executor) bool
	CurrentVersion(exec Executor) (string, error)
	LatestVersion(exec Executor) (string, error) // "" + nil means unknown, not an error
	Upgrade(exec Executor) error                 // idempotent: no-op if already current
}

// Evaluate determines the Status of a tool. Pure derivation: any error or empty
// version collapses to StateUnknown rather than failing the run.
func Evaluate(t Tool, exec Executor) Status {
	s := Status{Name: t.Name()}
	if !t.Installed(exec) {
		s.State = StateNotInstalled
		return s
	}
	cur, err := t.CurrentVersion(exec)
	if err != nil || cur == "" {
		s.State = StateUnknown
		return s
	}
	s.Current = cur

	latest, err := t.LatestVersion(exec)
	if err != nil || latest == "" {
		s.State = StateUnknown
		return s
	}
	s.Latest = latest

	if cur == latest {
		s.State = StateUpToDate
	} else {
		s.State = StateUpdateAvailable
	}
	return s
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd cli && go test ./internal/toolupgrade/ -run TestEvaluate -v`
Expected: PASS (4 subtests).

- [ ] **Step 5: Commit**

```bash
git add cli/internal/toolupgrade/toolupgrade.go cli/internal/toolupgrade/toolupgrade_test.go
git commit -m "feat(toolupgrade): add core types and Evaluate"
```

---

## Task 2: `Run` orchestrator + `StdinDecider`

Adds the check → prompt-all → apply flow. The `Decider` is invoked once with all candidates before any `Upgrade` call.

**Files:**
- Modify: `cli/internal/toolupgrade/toolupgrade.go` (append)
- Modify: `cli/internal/toolupgrade/toolupgrade_test.go` (append)

- [ ] **Step 1: Write the failing test**

Append to `cli/internal/toolupgrade/toolupgrade_test.go`:

```go
import_block_note: // also add "bytes" and "errors" and "strings" to the import list

func TestRunCheckModeRunsNothing(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0"}
	deciderCalled := false
	decider := func([]toolupgrade.Status) (map[string]bool, error) {
		deciderCalled = true
		return map[string]bool{"a": true}, nil
	}
	var buf bytes.Buffer
	err := toolupgrade.Run(
		toolupgrade.Options{Check: true, Out: &buf},
		nil,
		[]toolupgrade.Tool{a},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if deciderCalled {
		t.Error("decider must not be called in --check mode")
	}
	if a.upgraded {
		t.Error("no upgrade may run in --check mode")
	}
	if !strings.Contains(buf.String(), "update available") {
		t.Errorf("table missing status; got: %q", buf.String())
	}
}

func TestRunOnlyUpgradesApprovedAfterPrompting(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0"}
	b := &fakeTool{name: "b", installed: true, current: "2.0.0", latest: "2.5.0"}

	var sawCandidates []string
	decider := func(cands []toolupgrade.Status) (map[string]bool, error) {
		for _, c := range cands {
			sawCandidates = append(sawCandidates, c.Name)
		}
		// Verify "ask everything first": no upgrade has run yet.
		if a.upgraded || b.upgraded {
			t.Fatal("an upgrade ran before the decider returned")
		}
		return map[string]bool{"a": true, "b": false}, nil
	}
	var buf bytes.Buffer
	err := toolupgrade.Run(
		toolupgrade.Options{Out: &buf},
		nil,
		[]toolupgrade.Tool{a, b},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(sawCandidates) != 2 {
		t.Fatalf("decider should see 2 candidates, saw %v", sawCandidates)
	}
	if !a.upgraded {
		t.Error("approved tool a was not upgraded")
	}
	if b.upgraded {
		t.Error("unapproved tool b was upgraded")
	}
	if !strings.Contains(buf.String(), "1 upgraded, 1 skipped, 0 failed") {
		t.Errorf("summary wrong; got: %q", buf.String())
	}
}

func TestRunContinuesAfterFailure(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0", upgradeErr: errors.New("boom")}
	b := &fakeTool{name: "b", installed: true, current: "2.0.0", latest: "2.5.0"}
	decider := func([]toolupgrade.Status) (map[string]bool, error) {
		return map[string]bool{"a": true, "b": true}, nil
	}
	var buf bytes.Buffer
	if err := toolupgrade.Run(toolupgrade.Options{Out: &buf}, nil, []toolupgrade.Tool{a, b}, decider); err != nil {
		t.Fatal(err)
	}
	if !b.upgraded {
		t.Error("tool b should still upgrade after a fails")
	}
	if !strings.Contains(buf.String(), "1 upgraded, 0 skipped, 1 failed") {
		t.Errorf("summary wrong; got: %q", buf.String())
	}
}

func TestRunNoCandidatesSkipsDecider(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.0.0"} // up to date
	called := false
	decider := func([]toolupgrade.Status) (map[string]bool, error) { called = true; return nil, nil }
	var buf bytes.Buffer
	if err := toolupgrade.Run(toolupgrade.Options{Out: &buf}, nil, []toolupgrade.Tool{a}, decider); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("decider must not be called when nothing is upgradable")
	}
	if !strings.Contains(buf.String(), "All tools up to date") {
		t.Errorf("expected up-to-date message; got: %q", buf.String())
	}
}

func TestStdinDeciderCollectsAllAnswers(t *testing.T) {
	cands := []toolupgrade.Status{
		{Name: "a", Current: "1.0.0", Latest: "1.1.0"},
		{Name: "b", Current: "2.0.0", Latest: ""},
	}
	in := strings.NewReader("y\nn\n")
	var out bytes.Buffer
	decider := toolupgrade.StdinDecider(in, &out)
	approved, err := decider(cands)
	if err != nil {
		t.Fatal(err)
	}
	if !approved["a"] || approved["b"] {
		t.Errorf("approvals wrong: %v", approved)
	}
	if !strings.Contains(out.String(), "Upgrade b (2.0.0 → ?)") {
		t.Errorf("unknown-latest prompt should show '?'; got: %q", out.String())
	}
}
```

Then update the file's import block to:

```go
import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)
```

(Remove the `import_block_note:` line — it is only a reminder, not code.)

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/toolupgrade/ -run 'TestRun|TestStdinDecider' -v`
Expected: FAIL — `undefined: toolupgrade.Run`, `toolupgrade.Options`, `toolupgrade.StdinDecider`.

- [ ] **Step 3: Write minimal implementation**

Append to `cli/internal/toolupgrade/toolupgrade.go`, and add the imports shown below to its import block:

```go
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Options configures a Run.
type Options struct {
	Check bool      // print the table and exit; no prompts, no upgrades
	Out   io.Writer // defaults to os.Stdout if nil
}

// Decider receives every upgrade candidate and returns the set of tool names the
// user approved. It MUST collect all answers before returning; Run applies
// upgrades only after it returns, so no upgrade can run mid-prompt.
type Decider func(candidates []Status) (approved map[string]bool, err error)

// Run evaluates all tools, prints a status table, and (unless Check) prompts via
// decider and applies the approved upgrades.
func Run(opts Options, exec Executor, tools []Tool, decider Decider) error {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	statuses := make([]Status, 0, len(tools))
	byName := make(map[string]Tool, len(tools))
	for _, t := range tools {
		statuses = append(statuses, Evaluate(t, exec))
		byName[t.Name()] = t
	}
	renderTable(out, statuses)

	if opts.Check {
		return nil
	}

	var candidates []Status
	for _, s := range statuses {
		if s.State == StateUpdateAvailable || s.State == StateUnknown {
			candidates = append(candidates, s)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(out, "\nAll tools up to date.")
		return nil
	}

	approved, err := decider(candidates)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	var upgraded, skipped, failed int
	for _, c := range candidates {
		if !approved[c.Name] {
			skipped++
			continue
		}
		fmt.Fprintf(out, "  → %s … ", c.Name)
		if err := byName[c.Name].Upgrade(exec); err != nil {
			fmt.Fprintf(out, "FAILED: %v\n", err)
			failed++
			continue
		}
		fmt.Fprintln(out, "done")
		upgraded++
	}
	fmt.Fprintf(out, "\nSummary: %d upgraded, %d skipped, %d failed.\n", upgraded, skipped, failed)
	return nil
}

func renderTable(out io.Writer, statuses []Status) {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TOOL\tCURRENT\tLATEST\tSTATUS")
	for _, s := range statuses {
		cur := s.Current
		if cur == "" {
			cur = "—"
		}
		latest := s.Latest
		if latest == "" {
			latest = "—"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Name, cur, latest, stateLabel(s.State))
	}
	tw.Flush()
}

func stateLabel(st State) string {
	switch st {
	case StateUpToDate:
		return "up to date"
	case StateUpdateAvailable:
		return "update available"
	case StateNotInstalled:
		return "not installed"
	default:
		return "unknown"
	}
}

// StdinDecider returns a Decider that asks y/N for each candidate, reading from in
// and prompting to out. It reads all answers before returning. Empty/EOF ⇒ No.
func StdinDecider(in io.Reader, out io.Writer) Decider {
	return func(candidates []Status) (map[string]bool, error) {
		approved := make(map[string]bool, len(candidates))
		reader := bufio.NewReader(in)
		for _, c := range candidates {
			latest := c.Latest
			if latest == "" {
				latest = "?"
			}
			fmt.Fprintf(out, "Upgrade %s (%s → %s)? [y/N] ", c.Name, c.Current, latest)
			line, _ := reader.ReadString('\n')
			ans := strings.ToLower(strings.TrimSpace(line))
			approved[c.Name] = ans == "y" || ans == "yes"
		}
		return approved, nil
	}
}
```

Note: merge this import block with the one already at the top of the file from Task 1 (Task 1 added no imports, so this becomes the file's single import block).

- [ ] **Step 4: Run test to verify it passes**

Run: `cd cli && go test ./internal/toolupgrade/ -v`
Expected: PASS — all `TestEvaluate`, `TestRun*`, and `TestStdinDecider*` subtests green.

- [ ] **Step 5: Commit**

```bash
git add cli/internal/toolupgrade/toolupgrade.go cli/internal/toolupgrade/toolupgrade_test.go
git commit -m "feat(toolupgrade): add Run orchestrator and StdinDecider"
```

---

## Task 3: Concrete tools + `Registry` + helpers

Implements brew/claude/opencode against the verified real behavior, plus the shared parsing helpers.

**Files:**
- Create: `cli/internal/toolupgrade/tools.go`
- Create: `cli/internal/toolupgrade/tools_test.go`

- [ ] **Step 1: Write the failing test**

Create `cli/internal/toolupgrade/tools_test.go`:

```go
package toolupgrade_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)

// fakeExec keys Run results by the full command line ("cmd arg1 arg2") so the
// same binary (e.g. brew) can return different output per argument set.
type fakeExec struct {
	calls   [][]string
	outputs map[string]string
	errs    map[string]error
	paths   map[string]bool
}

func newFakeExec() *fakeExec {
	return &fakeExec{outputs: map[string]string{}, errs: map[string]error{}, paths: map[string]bool{}}
}

func cmdKey(cmd string, args []string) string {
	return strings.TrimSpace(cmd + " " + strings.Join(args, " "))
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string{cmd}, args...))
	k := cmdKey(cmd, args)
	return f.outputs[k], f.errs[k]
}

func (f *fakeExec) LookPath(name string) (string, error) {
	if f.paths[name] {
		return "/opt/homebrew/bin/" + name, nil
	}
	return "", errors.New("not found")
}

// findTool returns the registered tool with the given name.
func findTool(t *testing.T, name string) toolupgrade.Tool {
	t.Helper()
	for _, tool := range toolupgrade.Registry() {
		if tool.Name() == name {
			return tool
		}
	}
	t.Fatalf("tool %q not in registry", name)
	return nil
}

func TestRegistryContainsExpectedTools(t *testing.T) {
	got := map[string]bool{}
	for _, tool := range toolupgrade.Registry() {
		got[tool.Name()] = true
	}
	for _, want := range []string{"brew", "claude", "opencode"} {
		if !got[want] {
			t.Errorf("registry missing %q", want)
		}
	}
}

func TestBrewVersions(t *testing.T) {
	exec := newFakeExec()
	exec.paths["brew"] = true
	exec.outputs["brew --version"] = "Homebrew 5.1.14\nHomebrew/homebrew-core (git revision abc)"
	exec.outputs["curl -fsSL https://api.github.com/repos/Homebrew/brew/releases/latest"] = `{"tag_name":"5.1.15"}`

	brew := findTool(t, "brew")
	if !brew.Installed(exec) {
		t.Error("brew should be installed")
	}
	if cur, _ := brew.CurrentVersion(exec); cur != "5.1.14" {
		t.Errorf("current: want 5.1.14, got %q", cur)
	}
	if latest, _ := brew.LatestVersion(exec); latest != "5.1.15" {
		t.Errorf("latest: want 5.1.15, got %q", latest)
	}
}

func TestClaudeVersions(t *testing.T) {
	exec := newFakeExec()
	exec.paths["claude"] = true
	exec.outputs["claude --version"] = "2.1.165 (Claude Code)"

	claude := findTool(t, "claude")
	if cur, _ := claude.CurrentVersion(exec); cur != "2.1.165" {
		t.Errorf("current: want 2.1.165, got %q", cur)
	}
	latest, err := claude.LatestVersion(exec)
	if err != nil {
		t.Fatal(err)
	}
	if latest != "" {
		t.Errorf("claude latest should be unknown (\"\"), got %q", latest)
	}
}

func TestOpencodeVersionsStripsVPrefix(t *testing.T) {
	exec := newFakeExec()
	exec.paths["opencode"] = true
	exec.outputs["opencode --version"] = "1.2.10"
	exec.outputs["curl -fsSL https://api.github.com/repos/sst/opencode/releases/latest"] = `{"tag_name":"v1.16.0"}`

	oc := findTool(t, "opencode")
	if cur, _ := oc.CurrentVersion(exec); cur != "1.2.10" {
		t.Errorf("current: want 1.2.10, got %q", cur)
	}
	if latest, _ := oc.LatestVersion(exec); latest != "1.16.0" {
		t.Errorf("latest: want 1.16.0 (v stripped), got %q", latest)
	}
}

func TestNotInstalledWhenLookPathFails(t *testing.T) {
	exec := newFakeExec() // no paths set
	if findTool(t, "brew").Installed(exec) {
		t.Error("brew should report not installed when LookPath fails")
	}
}

func TestUpgradeCommands(t *testing.T) {
	cases := []struct {
		tool     string
		wantArgs []string
	}{
		{"brew", []string{"brew", "update"}},
		{"claude", []string{"claude", "update"}},
		{"opencode", []string{"opencode", "upgrade"}},
	}
	for _, c := range cases {
		t.Run(c.tool, func(t *testing.T) {
			exec := newFakeExec()
			if err := findTool(t, c.tool).Upgrade(exec); err != nil {
				t.Fatal(err)
			}
			if len(exec.calls) != 1 {
				t.Fatalf("want 1 call, got %d: %v", len(exec.calls), exec.calls)
			}
			got := exec.calls[0]
			if strings.Join(got, " ") != strings.Join(c.wantArgs, " ") {
				t.Errorf("want %v, got %v", c.wantArgs, got)
			}
		})
	}
}

func TestUpgradeWrapsError(t *testing.T) {
	exec := newFakeExec()
	exec.errs["brew update"] = errors.New("network down")
	if err := findTool(t, "brew").Upgrade(exec); err == nil {
		t.Fatal("expected error when brew update fails")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/toolupgrade/ -run 'TestRegistry|TestBrew|TestClaude|TestOpencode|TestNotInstalled|TestUpgrade' -v`
Expected: FAIL — `undefined: toolupgrade.Registry`.

- [ ] **Step 3: Write minimal implementation**

Create `cli/internal/toolupgrade/tools.go`:

```go
package toolupgrade

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Registry returns the ordered set of top-level tools to manage.
func Registry() []Tool {
	return []Tool{brewTool{}, claudeTool{}, opencodeTool{}}
}

// --- helpers ---

// firstLine returns the first line of s (without the trailing newline).
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// firstField returns the first whitespace-separated token of s.
func firstField(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// githubLatestTag fetches the latest release tag for owner/repo via the GitHub
// API and strips a leading "v". Returns "" only on parse of an empty tag.
func githubLatestTag(exec Executor, repo string) (string, error) {
	out, err := exec.Run("curl", "-fsSL", "https://api.github.com/repos/"+repo+"/releases/latest")
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal([]byte(out), &rel); err != nil {
		return "", fmt.Errorf("parse %s release: %w", repo, err)
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

// --- brew ---

type brewTool struct{}

func (brewTool) Name() string                  { return "brew" }
func (brewTool) Installed(exec Executor) bool   { _, err := exec.LookPath("brew"); return err == nil }
func (brewTool) CurrentVersion(exec Executor) (string, error) {
	out, err := exec.Run("brew", "--version")
	if err != nil {
		return "", err
	}
	// "Homebrew 5.1.14" → "5.1.14"
	fields := strings.Fields(firstLine(out))
	if len(fields) < 2 {
		return "", fmt.Errorf("unexpected brew --version output: %q", out)
	}
	return fields[1], nil
}
func (brewTool) LatestVersion(exec Executor) (string, error) {
	return githubLatestTag(exec, "Homebrew/brew")
}
func (brewTool) Upgrade(exec Executor) error {
	if _, err := exec.Run("brew", "update"); err != nil {
		return fmt.Errorf("brew update: %w", err)
	}
	return nil
}

// --- claude ---

type claudeTool struct{}

func (claudeTool) Name() string                { return "claude" }
func (claudeTool) Installed(exec Executor) bool { _, err := exec.LookPath("claude"); return err == nil }
func (claudeTool) CurrentVersion(exec Executor) (string, error) {
	out, err := exec.Run("claude", "--version")
	if err != nil {
		return "", err
	}
	// "2.1.165 (Claude Code)" → "2.1.165"
	return firstField(out), nil
}
func (claudeTool) LatestVersion(Executor) (string, error) {
	// No clean pre-check endpoint; `claude update` determines latest on apply.
	return "", nil
}
func (claudeTool) Upgrade(exec Executor) error {
	if _, err := exec.Run("claude", "update"); err != nil {
		return fmt.Errorf("claude update: %w", err)
	}
	return nil
}

// --- opencode ---

type opencodeTool struct{}

func (opencodeTool) Name() string                { return "opencode" }
func (opencodeTool) Installed(exec Executor) bool { _, err := exec.LookPath("opencode"); return err == nil }
func (opencodeTool) CurrentVersion(exec Executor) (string, error) {
	out, err := exec.Run("opencode", "--version")
	if err != nil {
		return "", err
	}
	// "1.2.10" → "1.2.10"
	return firstField(out), nil
}
func (opencodeTool) LatestVersion(exec Executor) (string, error) {
	return githubLatestTag(exec, "sst/opencode")
}
func (opencodeTool) Upgrade(exec Executor) error {
	if _, err := exec.Run("opencode", "upgrade"); err != nil {
		return fmt.Errorf("opencode upgrade: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd cli && go test ./internal/toolupgrade/ -v`
Expected: PASS — all tests across both test files green.

- [ ] **Step 5: Commit**

```bash
git add cli/internal/toolupgrade/tools.go cli/internal/toolupgrade/tools_test.go
git commit -m "feat(toolupgrade): add brew, claude, opencode tools and registry"
```

---

## Task 4: Wire up `bs tool upgrade` command

Adds the Cobra group, the `upgrade` subcommand with `--check`, and registers it on root.

**Files:**
- Create: `cli/cmd/tool/tool.go`
- Create: `cli/cmd/tool/upgrade.go`
- Modify: `cli/cmd/root.go`

- [ ] **Step 1: Create the group command file**

Create `cli/cmd/tool/tool.go`:

```go
package tool

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs tool' group command.
var Cmd = &cobra.Command{
	Use:   "tool",
	Short: "Manage top-level CLI tools not handled by brew or mise",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(upgradeCmd)
}

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *realExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
```

- [ ] **Step 2: Create the upgrade subcommand file**

Create `cli/cmd/tool/upgrade.go`:

```go
package tool

import (
	"os"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
	"github.com/spf13/cobra"
)

var checkOnly bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check and optionally upgrade top-level tools (brew, claude, opencode)",
	Long: `upgrade checks each top-level tool's installed version against the latest
available version, prompts yes/no for every out-of-date tool up front, then applies
only the approved upgrades.

Use --check to print the status table without prompting or upgrading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor := &realExecutor{}
		return toolupgrade.Run(
			toolupgrade.Options{Check: checkOnly, Out: os.Stdout},
			executor,
			toolupgrade.Registry(),
			toolupgrade.StdinDecider(os.Stdin, os.Stdout),
		)
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "print status and exit without prompting or upgrading")
}
```

- [ ] **Step 3: Register the group on root**

Modify `cli/cmd/root.go`. Add the import (keep alphabetical with the other `cmd/...` imports):

```go
	"github.com/dlstadther/bootstrap/cli/cmd/tool"
```

And add to the `init()` func, after `rootCmd.AddCommand(tmux.Cmd)`:

```go
	rootCmd.AddCommand(tool.Cmd)
```

- [ ] **Step 4: Build and verify the command is wired**

Run: `cd cli && go build ./... && go run . tool upgrade --help`
Expected: build succeeds; help text shows `Check and optionally upgrade top-level tools (brew, claude, opencode)` and the `--check` flag.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/tool/tool.go cli/cmd/tool/upgrade.go cli/cmd/root.go
git commit -m "feat(tool): wire up bs tool upgrade command"
```

---

## Task 5: Full verification + real smoke test

Confirms the whole package and a real `--check` run against live tools.

**Files:** none (verification only)

- [ ] **Step 1: Run the full Go test suite**

Run: `cd cli && go test ./...`
Expected: all packages PASS (including pre-existing `internal/brew`, `internal/audit`, etc.).

- [ ] **Step 2: Vet the code**

Run: `cd cli && go vet ./...`
Expected: no output (clean).

- [ ] **Step 3: Real check-only smoke test**

Run: `cd cli && go run . tool upgrade --check`
Expected: a table listing `brew`, `claude`, `opencode` with real CURRENT values, a real LATEST for brew + opencode (claude shows `—`), and a STATUS per tool. No prompts appear, no upgrade runs.

- [ ] **Step 4: Real interactive smoke test answering "no" to all**

Run: `cd cli && printf 'n\nn\nn\n' | go run . tool upgrade`
Expected: the table prints, then `Upgrade …? [y/N]` lines, then `Summary: 0 upgraded, N skipped, 0 failed.` No upgrade commands run.

- [ ] **Step 5: Install and final commit**

Run: `make install` (rebuilds `bs` to `~/.local/bin/bs` and re-symlinks dotfiles), then `bs tool upgrade --check`.
Expected: same table as Step 3, now via the installed binary.

```bash
git commit --allow-empty -m "test(tool): verify bs tool upgrade end-to-end"
```

---

## Self-Review

**Spec coverage:**
- brew/claude/opencode in scope → Task 3 ✓
- check all versions (current vs latest) → `Evaluate` + tools (Tasks 1, 3) ✓
- prompt y/n for every tool BEFORE any upgrade → `Run` + `Decider`, asserted by `TestRunOnlyUpgradesApprovedAfterPrompting` (Task 2) ✓
- apply only approved → `Run`, asserted (Task 2) ✓
- `--check` non-interactive mode → `Options.Check` + flag (Tasks 2, 4), asserted `TestRunCheckModeRunsNothing` + smoke Step 3 ✓
- error handling: not installed (warn/skip) → `Evaluate` StateNotInstalled (Task 1); latest fetch fails → Unknown, still opt-in (Task 1/3); upgrade fails → continue + count → `TestRunContinuesAfterFailure` (Task 2); EOF ⇒ No → `StdinDecider` (Task 2) ✓
- status table output → `renderTable` (Task 2) ✓
- registry extensibility → `Registry()` (Task 3) ✓
- YAGNI: no `--yes`, no pin list, no make target → none added ✓

**Placeholder scan:** No TBD/TODO. The `import_block_note:` token in Task 2 Step 1 is explicitly flagged as a non-code reminder to remove. All code steps contain complete code.

**Type consistency:** `Executor` (Run + LookPath), `Tool` (Name/Installed/CurrentVersion/LatestVersion/Upgrade), `Status{Name,Current,Latest,State}`, `State` constants (`StateUpToDate/StateUpdateAvailable/StateUnknown/StateNotInstalled`), `Options{Check,Out}`, `Decider`, `Evaluate`, `Run`, `StdinDecider`, `Registry` — names and signatures are identical across the spec, tests, and implementations in Tasks 1–4. The `fakeExec` keys Run by full command line, matching how tools call `exec.Run` (e.g. `"brew --version"`, `"curl -fsSL …"`).
