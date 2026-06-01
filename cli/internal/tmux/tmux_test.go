package tmux_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/tmux"
)

type callRecord struct {
	cmd  string
	args []string
}

type fakeExec struct {
	calls   []callRecord
	results map[string]string // keyed by first arg after cmd, e.g. "has-session" -> output
	errs    map[string]error
}

func newFake() *fakeExec {
	return &fakeExec{results: map[string]string{}, errs: map[string]error{}}
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
	err := tmux.Add(tmux.AddOptions{Agent: "claude"}, exec)
	if err == nil || !strings.Contains(err.Error(), "--cwd") {
		t.Fatalf("expected --cwd error, got %v", err)
	}
}

func TestAdd_InvalidAgent(t *testing.T) {
	exec := newFake()
	err := tmux.Add(tmux.AddOptions{CWD: "/some/path", Agent: "badagent"}, exec)
	if err == nil || !strings.Contains(err.Error(), "invalid agent") {
		t.Fatalf("expected invalid agent error, got %v", err)
	}
}

func TestAdd_TmuxNotRunning(t *testing.T) {
	exec := newFake()
	exec.errs["tmux info"] = errors.New("no server running")
	err := tmux.Add(tmux.AddOptions{CWD: "/some/path", Agent: "claude"}, exec)
	if err == nil || !strings.Contains(err.Error(), "tmux is not running") {
		t.Fatalf("expected tmux not running error, got %v", err)
	}
}

func TestAdd_NewSession(t *testing.T) {
	exec := newFake()
	// tmux info succeeds (tmux is running)
	// has-session fails → no existing session
	exec.errs["tmux has-session"] = errors.New("no session")

	err := tmux.Add(tmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	// Verify new-session was called
	found := false
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 && c.args[0] == "new-session" {
			found = true
		}
	}
	if !found {
		t.Error("expected new-session call, not found")
	}
}

func TestAdd_ExistingSession(t *testing.T) {
	exec := newFake()
	// has-session succeeds → session exists
	// list-windows returns no matching window
	exec.results["tmux list-windows"] = "other-window\n"

	err := tmux.Add(tmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 && c.args[0] == "new-window" {
			found = true
		}
	}
	if !found {
		t.Error("expected new-window call, not found")
	}
}

func TestAdd_AgentStagedUsingPaneID(t *testing.T) {
	exec := newFake()
	exec.errs["tmux has-session"] = errors.New("no session")
	exec.results["tmux display-message"] = "%42"

	err := tmux.Add(tmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	// Verify send-keys for the agent uses the pane ID, not a hardcoded .0 index.
	agentSent := false
	for _, c := range exec.calls {
		if c.cmd != "tmux" || len(c.args) < 3 || c.args[0] != "send-keys" {
			continue
		}
		// Find the send-keys call whose value is "claude" (no Enter = staging)
		for i, arg := range c.args {
			if arg == "claude" && i+1 < len(c.args) && c.args[i+1] == "" {
				// The -t arg should be the pane ID, not session:window.0
				for j, a := range c.args {
					if a == "-t" && j+1 < len(c.args) {
						if c.args[j+1] == "%42" {
							agentSent = true
						}
					}
				}
			}
		}
	}
	if !agentSent {
		t.Error("expected send-keys for agent to target pane ID %42")
	}
}

func TestAdd_AgentStagedFallsBackWhenNoPaneID(t *testing.T) {
	exec := newFake()
	exec.errs["tmux has-session"] = errors.New("no session")
	// display-message returns empty — fallback to session:window target

	err := tmux.Add(tmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	agentSent := false
	for _, c := range exec.calls {
		if c.cmd != "tmux" || len(c.args) < 3 || c.args[0] != "send-keys" {
			continue
		}
		for i, arg := range c.args {
			if arg == "claude" && i+1 < len(c.args) && c.args[i+1] == "" {
				for j, a := range c.args {
					if a == "-t" && j+1 < len(c.args) {
						if c.args[j+1] == "myproject:myproject" {
							agentSent = true
						}
					}
				}
			}
		}
	}
	if !agentSent {
		t.Error("expected send-keys for agent to fall back to session:window target")
	}
}

func TestAllowedAgents(t *testing.T) {
	expected := []string{"claude", "codex", "gemini", "opencode", "pi"}
	if len(tmux.AllowedAgents) != len(expected) {
		t.Fatalf("expected %d agents, got %d", len(expected), len(tmux.AllowedAgents))
	}
	for i, a := range expected {
		if tmux.AllowedAgents[i] != a {
			t.Errorf("agent[%d]: want %s got %s", i, a, tmux.AllowedAgents[i])
		}
	}
}
