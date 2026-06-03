package cmux_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/cmux"
)

type callRecord struct {
	cmd  string
	args []string
}

type fakeExec struct {
	calls   []callRecord
	results map[string]string // keyed by "cmd subcmd", e.g. "cmux ping"
	errs    map[string]error
}

func newFake() *fakeExec {
	f := &fakeExec{results: map[string]string{}, errs: map[string]error{}}
	// cmux ping succeeds by default (cmux is running)
	f.results["cmux new-workspace"] = "workspace:1"
	f.results["cmux list-panes"] = "pane:1"
	return f
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, callRecord{cmd: cmd, args: args})
	key := cmd
	if len(args) > 0 {
		key = cmd + " " + args[0]
	}
	return f.results[key], f.errs[key]
}

func TestAdd_MissingCWD(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{Agent: "claude"}, exec)
	if err == nil || !strings.Contains(err.Error(), "--cwd") {
		t.Fatalf("expected --cwd error, got %v", err)
	}
}

func TestAdd_InvalidAgent(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/some/path", Agent: "badagent"}, exec)
	if err == nil || !strings.Contains(err.Error(), "invalid agent") {
		t.Fatalf("expected invalid agent error, got %v", err)
	}
}

func TestAdd_CmuxNotRunning(t *testing.T) {
	exec := newFake()
	exec.errs["cmux ping"] = errors.New("connection refused")
	err := cmux.Add(cmux.AddOptions{CWD: "/some/path", Agent: "claude"}, exec)
	if err == nil || !strings.Contains(err.Error(), "cmux is not running") {
		t.Fatalf("expected cmux not running error, got %v", err)
	}
}

func TestAdd_NewWorkspace(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-workspace" {
			found = true
			// Verify --name defaults to basename of CWD
			nameIdx := indexOf(c.args, "--name")
			if nameIdx >= 0 && c.args[nameIdx+1] != "myproject" {
				t.Errorf("expected --name myproject, got %s", c.args[nameIdx+1])
			}
			// Verify --cwd is passed
			cwdIdx := indexOf(c.args, "--cwd")
			if cwdIdx < 0 {
				t.Error("expected --cwd flag in new-workspace call")
			}
		}
	}
	if !found {
		t.Error("expected new-workspace call, not found")
	}
}

func TestAdd_WorkspaceNameOverride(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Name: "custom", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-workspace" {
			nameIdx := indexOf(c.args, "--name")
			if nameIdx >= 0 && c.args[nameIdx+1] != "custom" {
				t.Errorf("expected --name custom, got %s", c.args[nameIdx+1])
			}
		}
	}
}

func TestAdd_SplitsCreated(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	var splits []string
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-split" {
			if len(c.args) > 1 {
				splits = append(splits, c.args[1])
			}
		}
	}
	if len(splits) != 2 {
		t.Fatalf("expected 2 new-split calls, got %d", len(splits))
	}
	if splits[0] != "right" {
		t.Errorf("expected first split to be right, got %s", splits[0])
	}
	if splits[1] != "down" {
		t.Errorf("expected second split to be down, got %s", splits[1])
	}
}

func TestAdd_ClaudeAgentStaged(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	// Find the last send call — it should be the staged agent command.
	var lastSend callRecord
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "send" {
			lastSend = c
		}
	}
	if lastSend.cmd == "" {
		t.Fatal("expected at least one send call")
	}
	text := lastSend.args[len(lastSend.args)-1]
	if !strings.Contains(text, "claude agents") || !strings.Contains(text, "/code/myproject") {
		t.Errorf("expected staged claude agents command, got %q", text)
	}

	// Verify no send-key enter follows the staged command.
	// The last send-key call should precede the final send (lazygit's enter).
	var sendKeyCalls []callRecord
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "send-key" {
			sendKeyCalls = append(sendKeyCalls, c)
		}
	}
	// Should be exactly 2 send-key enter calls (top-right and bottom-right), not 3.
	enterCount := 0
	for _, c := range sendKeyCalls {
		if c.args[len(c.args)-1] == "enter" {
			enterCount++
		}
	}
	if enterCount != 2 {
		t.Errorf("expected 2 send-key enter calls (not 3), got %d", enterCount)
	}
}

func TestAdd_NonClaudeAgent(t *testing.T) {
	exec := newFake()
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "codex"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	var lastSend callRecord
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "send" {
			lastSend = c
		}
	}
	text := lastSend.args[len(lastSend.args)-1]
	if text != "codex" {
		t.Errorf("expected staged command to be bare agent name %q, got %q", "codex", text)
	}
}

func TestAdd_FocusLeftPane(t *testing.T) {
	exec := newFake()
	exec.results["cmux list-panes"] = "pane:42"
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "focus-pane" {
			paneIdx := indexOf(c.args, "--pane")
			if paneIdx >= 0 && c.args[paneIdx+1] == "pane:42" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected focus-pane call targeting pane:42")
	}
}

func TestAllowedAgents(t *testing.T) {
	expected := []string{"claude", "codex", "gemini", "opencode", "pi"}
	if len(cmux.AllowedAgents) != len(expected) {
		t.Fatalf("expected %d agents, got %d", len(expected), len(cmux.AllowedAgents))
	}
	for i, a := range expected {
		if cmux.AllowedAgents[i] != a {
			t.Errorf("agent[%d]: want %s got %s", i, a, cmux.AllowedAgents[i])
		}
	}
}

func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
