# bs claude plugin upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `bs claude plugin upgrade` to bulk-upgrade all enabled Claude Code plugins with the same check/prompt/apply UX as `bs tool upgrade`.

**Architecture:** New `cli/internal/pluginupgrade` package mirrors `toolupgrade` with self-contained types; `Discover()` parses `claude plugins list` output to build a dynamic `[]Tool` slice; a new `cli/cmd/claude` command tree wires it together and is registered in `root.go`.

**Tech Stack:** Go 1.26, Cobra, `os/exec`, standard library only.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `cli/internal/pluginupgrade/pluginupgrade.go` | Create | State/Status/Tool/Executor/Decider types; Evaluate, Run, StdinDecider, renderTable |
| `cli/internal/pluginupgrade/pluginupgrade_test.go` | Create | Tests for Evaluate and Run |
| `cli/internal/pluginupgrade/plugins.go` | Create | Plugin struct + Discover() + ParsePluginList() |
| `cli/internal/pluginupgrade/plugins_test.go` | Create | Tests for ParsePluginList and Discover |
| `cli/cmd/claude/claude.go` | Create | `bs claude` group command + realExecutor |
| `cli/cmd/claude/plugin.go` | Create | `bs claude plugin` group command |
| `cli/cmd/claude/plugin_upgrade.go` | Create | `bs claude plugin upgrade` command with --check flag |
| `cli/cmd/root.go` | Modify | Register `claude.Cmd` |

---

### Task 1: pluginupgrade core package (TDD)

**Files:**
- Create: `cli/internal/pluginupgrade/pluginupgrade_test.go`
- Create: `cli/internal/pluginupgrade/pluginupgrade.go`

- [ ] **Step 1: Write the failing tests**

Create `cli/internal/pluginupgrade/pluginupgrade_test.go`:

```go
package pluginupgrade_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
)

// stubExec records calls and returns configured outputs/errors.
type stubExec struct {
	outputs map[string]string
	errs    map[string]error
	called  []string
}

func (s *stubExec) Run(cmd string, args ...string) (string, error) {
	key := strings.Join(append([]string{cmd}, args...), " ")
	s.called = append(s.called, key)
	if e := s.errs[key]; e != nil {
		return "", e
	}
	return s.outputs[key], nil
}

func (s *stubExec) LookPath(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

// stubTool is a configurable pluginupgrade.Tool for tests.
type stubTool struct {
	name       string
	installed  bool
	current    string
	curErr     error
	latest     string
	latErr     error
	upgradeErr error
}

func (t stubTool) Name() string                                            { return t.name }
func (t stubTool) Installed(_ pluginupgrade.Executor) bool                 { return t.installed }
func (t stubTool) CurrentVersion(_ pluginupgrade.Executor) (string, error) { return t.current, t.curErr }
func (t stubTool) LatestVersion(_ pluginupgrade.Executor) (string, error)  { return t.latest, t.latErr }
func (t stubTool) Upgrade(_ pluginupgrade.Executor) error                  { return t.upgradeErr }

func newExec() *stubExec {
	return &stubExec{outputs: map[string]string{}, errs: map[string]error{}}
}

// --- Evaluate ---

func TestEvaluate_NotInstalled(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: false}, newExec())
	if s.State != pluginupgrade.StateNotInstalled {
		t.Fatalf("got %v, want StateNotInstalled", s.State)
	}
}

func TestEvaluate_CurrentVersionError(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, curErr: errors.New("fail")}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_EmptyCurrentVersion(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: ""}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_UnknownLatest(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: ""}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_UpToDate(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: "1.0.0"}, newExec())
	if s.State != pluginupgrade.StateUpToDate {
		t.Fatalf("got %v, want StateUpToDate", s.State)
	}
}

func TestEvaluate_UpdateAvailable(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: "2.0.0"}, newExec())
	if s.State != pluginupgrade.StateUpdateAvailable {
		t.Fatalf("got %v, want StateUpdateAvailable", s.State)
	}
	if s.Current != "1.0.0" || s.Latest != "2.0.0" {
		t.Fatalf("unexpected versions: current=%q latest=%q", s.Current, s.Latest)
	}
}

// --- Run ---

func TestRun_CheckOnly_NoDeciderCall(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a", installed: true, current: "1.0.0"}
	calledDecider := false
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		calledDecider = true
		return nil, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Check: true, Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if calledDecider {
		t.Fatal("decider must not be called in check mode")
	}
	if !strings.Contains(out.String(), "plug-a") {
		t.Fatalf("output should contain plugin name, got: %s", out.String())
	}
}

func TestRun_AllUpToDate_NoCandidates(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a", installed: true, current: "1.0.0", latest: "1.0.0"}
	calledDecider := false
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		calledDecider = true
		return nil, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if calledDecider {
		t.Fatal("decider must not be called when nothing to upgrade")
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Fatalf("expected 'up to date' message, got: %s", out.String())
	}
}

func TestRun_ApprovedUpgrade_PrintsDone(t *testing.T) {
	var out bytes.Buffer
	e := newExec()
	e.outputs["claude plugins update plug-a@mp"] = ""
	tool := stubTool{name: "plug-a@mp", installed: true, current: "1.0.0"}
	decider := pluginupgrade.Decider(func(candidates []pluginupgrade.Status) (map[string]bool, error) {
		approved := map[string]bool{}
		for _, c := range candidates {
			approved[c.Name] = true
		}
		return approved, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		e,
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "done") {
		t.Fatalf("expected 'done' in output, got: %s", output)
	}
	if !strings.Contains(output, "1 upgraded") {
		t.Fatalf("expected '1 upgraded' in summary, got: %s", output)
	}
}

func TestRun_SkippedUpgrade_PrintsSummary(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a@mp", installed: true, current: "1.0.0"}
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		return map[string]bool{}, nil // approve nothing
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1 skipped") {
		t.Fatalf("expected '1 skipped', got: %s", out.String())
	}
}

func TestRun_UpgradeFailure_PrintsFailed(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{
		name:       "plug-a@mp",
		installed:  true,
		current:    "1.0.0",
		upgradeErr: errors.New("network error"),
	}
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		return map[string]bool{"plug-a@mp": true}, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1 failed") {
		t.Fatalf("expected '1 failed', got: %s", out.String())
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure (package doesn't exist yet)**

```bash
cd cli && go test ./internal/pluginupgrade/...
```

Expected: `cannot find package "github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"`

- [ ] **Step 3: Create pluginupgrade.go**

Create `cli/internal/pluginupgrade/pluginupgrade.go`:

```go
// Package pluginupgrade checks and upgrades installed Claude Code plugins.
package pluginupgrade

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
	LookPath(name string) (string, error)
}

// State is the upgrade status of a plugin.
type State int

const (
	StateUpToDate State = iota
	StateUpdateAvailable
	StateUnknown
	StateNotInstalled
)

// Status is the evaluated state of one plugin.
type Status struct {
	Name    string
	Current string
	Latest  string
	State   State
}

// Tool is one managed plugin.
type Tool interface {
	Name() string
	Installed(exec Executor) bool
	CurrentVersion(exec Executor) (string, error)
	LatestVersion(exec Executor) (string, error)
	Upgrade(exec Executor) error
}

// Evaluate determines the Status of a tool.
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

// Options configures a Run.
type Options struct {
	Check bool
	Out   io.Writer
}

// Decider receives every upgrade candidate and returns the approved set.
type Decider func(candidates []Status) (approved map[string]bool, err error)

// Run evaluates all tools, prints a status table, and (unless Check) prompts
// via decider and applies approved upgrades.
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
		fmt.Fprintln(out, "\nAll plugins up to date.")
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
	fmt.Fprintln(tw, "PLUGIN\tCURRENT\tLATEST\tSTATUS")
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

// StdinDecider returns a Decider that prompts y/N for each candidate.
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

- [ ] **Step 4: Run tests — expect pass**

```bash
cd cli && go test ./internal/pluginupgrade/... -v -run "TestEvaluate|TestRun"
```

Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/pluginupgrade/pluginupgrade.go cli/internal/pluginupgrade/pluginupgrade_test.go
git commit -m "feat(claude-plugin-upgrade): add pluginupgrade core package"
```

---

### Task 2: Plugin discovery (TDD)

**Files:**
- Create: `cli/internal/pluginupgrade/plugins_test.go`
- Create: `cli/internal/pluginupgrade/plugins.go`

- [ ] **Step 1: Write the failing tests**

Create `cli/internal/pluginupgrade/plugins_test.go`:

```go
package pluginupgrade_test

import (
	"fmt"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
)

const samplePluginList = `Installed plugins:

  ❯ agentsmith@dlstadther-agentsmith
    Version: 1.0.0
    Scope: user
    Status: ✔ enabled

  ❯ beads@local
    Version: 1.0.0
    Scope: user
    Status: ✘ disabled

  ❯ superpowers@claude-plugins-official
    Version: 5.1.0
    Scope: user
    Status: ✔ enabled

  ❯ code-review@claude-plugins-official
    Version: unknown
    Scope: user
    Status: ✔ enabled

  ❯ context7@claude-plugins-official
    Version: unknown
    Scope: user
    Status: ✘ disabled`

func TestParsePluginList_OnlyEnabled(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	if len(tools) != 3 {
		t.Fatalf("expected 3 enabled plugins, got %d", len(tools))
	}
}

func TestParsePluginList_Names(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name()] = true
	}
	want := []string{
		"agentsmith@dlstadther-agentsmith",
		"superpowers@claude-plugins-official",
		"code-review@claude-plugins-official",
	}
	for _, w := range want {
		if !names[w] {
			t.Errorf("missing plugin %q", w)
		}
	}
	if names["beads@local"] {
		t.Error("disabled plugin beads@local must not be included")
	}
	if names["context7@claude-plugins-official"] {
		t.Error("disabled plugin context7@claude-plugins-official must not be included")
	}
}

func TestParsePluginList_KnownVersion(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	for _, tool := range tools {
		if tool.Name() == "agentsmith@dlstadther-agentsmith" {
			ver, err := tool.CurrentVersion(nil)
			if err != nil {
				t.Fatal(err)
			}
			if ver != "1.0.0" {
				t.Fatalf("expected 1.0.0, got %q", ver)
			}
			return
		}
	}
	t.Fatal("agentsmith plugin not found")
}

func TestParsePluginList_UnknownVersion_ReturnsEmpty(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	for _, tool := range tools {
		if tool.Name() == "code-review@claude-plugins-official" {
			ver, err := tool.CurrentVersion(nil)
			if err != nil {
				t.Fatal(err)
			}
			if ver != "" {
				t.Fatalf("expected empty string for unknown version, got %q", ver)
			}
			return
		}
	}
	t.Fatal("code-review plugin not found")
}

func TestParsePluginList_Empty(t *testing.T) {
	tools := pluginupgrade.ParsePluginList("")
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestDiscover_RunsClaudePluginsList(t *testing.T) {
	e := &stubExec{
		outputs: map[string]string{
			"claude plugins list": samplePluginList,
		},
		errs: map[string]error{},
	}
	tools, err := pluginupgrade.Discover(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	called := false
	for _, c := range e.called {
		if c == "claude plugins list" {
			called = true
		}
	}
	if !called {
		t.Fatal("expected 'claude plugins list' to be called")
	}
}

func TestDiscover_ErrorPropagates(t *testing.T) {
	e := &stubExec{
		outputs: map[string]string{},
		errs:    map[string]error{"claude plugins list": fmt.Errorf("fake error")},
	}
	_, err := pluginupgrade.Discover(e)
	if err == nil {
		t.Fatal("expected error from Discover when exec fails")
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
cd cli && go test ./internal/pluginupgrade/... -run TestParsePluginList
```

Expected: `undefined: pluginupgrade.ParsePluginList`

- [ ] **Step 3: Create plugins.go**

Create `cli/internal/pluginupgrade/plugins.go`:

```go
package pluginupgrade

import (
	"fmt"
	"strings"
)

// Plugin is an installed, enabled Claude Code plugin.
type Plugin struct {
	name    string
	version string // empty when version reported as "unknown"
}

func (p Plugin) Name() string                              { return p.name }
func (p Plugin) Installed(_ Executor) bool                 { return true }
func (p Plugin) CurrentVersion(_ Executor) (string, error) { return p.version, nil }
func (p Plugin) LatestVersion(_ Executor) (string, error)  { return "", nil }
func (p Plugin) Upgrade(exec Executor) error {
	if _, err := exec.Run("claude", "plugins", "update", p.name); err != nil {
		return fmt.Errorf("claude plugins update %s: %w", p.name, err)
	}
	return nil
}

// Discover runs `claude plugins list` and returns only enabled plugins.
func Discover(exec Executor) ([]Tool, error) {
	out, err := exec.Run("claude", "plugins", "list")
	if err != nil {
		return nil, fmt.Errorf("claude plugins list: %w", err)
	}
	return ParsePluginList(out), nil
}

// ParsePluginList parses the output of `claude plugins list` and returns only
// enabled plugins. Exported for testing.
func ParsePluginList(output string) []Tool {
	var tools []Tool
	var cur *Plugin

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "❯ ") {
			cur = &Plugin{name: strings.TrimPrefix(line, "❯ ")}
			continue
		}
		if cur == nil {
			continue
		}
		if strings.HasPrefix(line, "Version: ") {
			v := strings.TrimPrefix(line, "Version: ")
			if v != "unknown" {
				cur.version = v
			}
			continue
		}
		if line == "Status: ✔ enabled" {
			tools = append(tools, *cur)
		}
	}
	return tools
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd cli && go test ./internal/pluginupgrade/... -v
```

Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/pluginupgrade/plugins.go cli/internal/pluginupgrade/plugins_test.go
git commit -m "feat(claude-plugin-upgrade): add plugin discovery with ParsePluginList"
```

---

### Task 3: Command tree

**Files:**
- Create: `cli/cmd/claude/claude.go`
- Create: `cli/cmd/claude/plugin.go`
- Create: `cli/cmd/claude/plugin_upgrade.go`

- [ ] **Step 1: Create claude.go (group command + shared realExecutor)**

Create `cli/cmd/claude/claude.go`:

```go
package claude

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs claude' group command.
var Cmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude Code configuration and plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(pluginCmd)
}

type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *realExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
```

- [ ] **Step 2: Create plugin.go (group command)**

Create `cli/cmd/claude/plugin.go`:

```go
package claude

import "github.com/spf13/cobra"

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Claude Code plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	pluginCmd.AddCommand(pluginUpgradeCmd)
}
```

- [ ] **Step 3: Create plugin_upgrade.go**

Create `cli/cmd/claude/plugin_upgrade.go`:

```go
package claude

import (
	"fmt"
	"os"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
	"github.com/spf13/cobra"
)

var pluginCheckOnly bool

var pluginUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check and optionally upgrade Claude Code plugins",
	Long: `upgrade lists all installed, enabled plugins, prompts yes/no for each
out-of-date or unknown-version plugin up front, then applies only the approved
upgrades.

Use --check to print the status table without prompting or upgrading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor := &realExecutor{}
		tools, err := pluginupgrade.Discover(executor)
		if err != nil {
			return err
		}
		if len(tools) == 0 {
			fmt.Fprintln(os.Stdout, "No enabled plugins found.")
			return nil
		}
		return pluginupgrade.Run(
			pluginupgrade.Options{Check: pluginCheckOnly, Out: os.Stdout},
			executor,
			tools,
			pluginupgrade.StdinDecider(os.Stdin, os.Stdout),
		)
	},
}

func init() {
	pluginUpgradeCmd.Flags().BoolVar(&pluginCheckOnly, "check", false, "print status and exit without prompting or upgrading")
}
```

- [ ] **Step 4: Build to verify no compile errors**

```bash
cd cli && go build ./cmd/claude/...
```

Expected: exits 0 with no output.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/claude/
git commit -m "feat(claude-plugin-upgrade): add bs claude plugin upgrade command tree"
```

---

### Task 4: Wire into root and verify end-to-end

**Files:**
- Modify: `cli/cmd/root.go`

- [ ] **Step 1: Add claude import and register Cmd in root.go**

In `cli/cmd/root.go`, the current import block is:

```go
import (
	"github.com/dlstadther/bootstrap/cli/cmd/brew"
	"github.com/dlstadther/bootstrap/cli/cmd/cmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tool"
	"github.com/spf13/cobra"
)
```

Replace it with:

```go
import (
	"github.com/dlstadther/bootstrap/cli/cmd/brew"
	"github.com/dlstadther/bootstrap/cli/cmd/claude"
	"github.com/dlstadther/bootstrap/cli/cmd/cmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tool"
	"github.com/spf13/cobra"
)
```

The current `init()` function is:

```go
func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(brew.Cmd)
	rootCmd.AddCommand(cmux.Cmd)
	rootCmd.AddCommand(tmux.Cmd)
	rootCmd.AddCommand(tool.Cmd)
}
```

Replace it with:

```go
func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(brew.Cmd)
	rootCmd.AddCommand(claude.Cmd)
	rootCmd.AddCommand(cmux.Cmd)
	rootCmd.AddCommand(tmux.Cmd)
	rootCmd.AddCommand(tool.Cmd)
}
```

- [ ] **Step 2: Full build**

```bash
cd cli && go build ./...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Run all tests**

```bash
cd cli && go test ./...
```

Expected: all packages PASS.

- [ ] **Step 4: Smoke test the command tree**

```bash
cd cli && go run . claude --help
```

Expected output includes:
```
Manage Claude Code configuration and plugins

Usage:
  bs claude [command]

Available Commands:
  plugin      Manage Claude Code plugins
```

```bash
cd cli && go run . claude plugin --help
```

Expected output includes:
```
Manage Claude Code plugins

Usage:
  bs claude plugin [command]

Available Commands:
  upgrade     Check and optionally upgrade Claude Code plugins
```

```bash
cd cli && go run . claude plugin upgrade --check
```

Expected: prints a plugin status table listing enabled plugins, exits 0.

- [ ] **Step 5: Install and verify installed binary**

```bash
make install && bs claude plugin upgrade --check
```

Expected: same status table from the installed binary.

- [ ] **Step 6: Commit**

```bash
git add cli/cmd/root.go
git commit -m "chore(claude-plugin-upgrade): register bs claude command in root"
```
