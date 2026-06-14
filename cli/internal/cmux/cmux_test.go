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
	f.results["cmux workspace create"] = "OK workspace:1"
	f.results["cmux list-panes"] = "* pane:1  [1 surface]  [focused]"
	f.results["cmux list-pane-surfaces pane:1"] = "* surface:1  Terminal  [selected]"
	return f
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, callRecord{cmd: cmd, args: args})
	key := cmd
	if len(args) > 0 {
		key = cmd + " " + args[0]
		if args[0] == "workspace" && len(args) > 1 {
			key = key + " " + args[1]
		}
		// list-pane-surfaces keys include --pane value for per-pane results.
		if args[0] == "list-pane-surfaces" {
			if pi := indexOf(args, "--pane"); pi >= 0 && pi+1 < len(args) {
				key = key + " " + args[pi+1]
			}
		}
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
		if c.cmd == "cmux" && len(c.args) > 1 && c.args[0] == "workspace" && c.args[1] == "create" {
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
		if c.cmd == "cmux" && len(c.args) > 1 && c.args[0] == "workspace" && c.args[1] == "create" {
			nameIdx := indexOf(c.args, "--name")
			if nameIdx >= 0 && c.args[nameIdx+1] != "custom" {
				t.Errorf("expected --name custom, got %s", c.args[nameIdx+1])
			}
		}
	}
}

func TestAdd_LayoutCreated(t *testing.T) {
	exec := newFake()
	exec.results["cmux list-panes"] = "* pane:1  [1 surface]  [focused]\n  pane:2  [1 surface]\n  pane:3  [1 surface]"
	exec.results["cmux list-pane-surfaces pane:1"] = "* surface:1  Terminal  [selected]"
	exec.results["cmux list-pane-surfaces pane:2"] = "* surface:2  Terminal  [selected]"
	exec.results["cmux list-pane-surfaces pane:3"] = "* surface:3  Terminal  [selected]"
	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err != nil {
		t.Fatal(err)
	}

	// Verify workspace create is called with --layout (not new-split).
	layoutFound := false
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 1 && c.args[0] == "workspace" && c.args[1] == "create" {
			if idx := indexOf(c.args, "--layout"); idx >= 0 {
				layoutFound = true
				if !strings.Contains(c.args[idx+1], "horizontal") {
					t.Errorf("expected layout JSON to contain horizontal split, got %s", c.args[idx+1])
				}
			}
		}
	}
	if !layoutFound {
		t.Error("expected workspace create with --layout flag")
	}

	// Verify no new-split calls are made.
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-split" {
			t.Error("unexpected new-split call; Add should use --layout instead")
		}
	}

	// Verify sends use surface refs (not pane refs) for targeting.
	surfaceCmds := map[string]string{}
	for _, c := range exec.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "send" {
			si := indexOf(c.args, "--surface")
			if si >= 0 && si+1 < len(c.args) {
				surfaceCmds[c.args[si+1]] = c.args[len(c.args)-1]
			}
		}
	}
	if !strings.Contains(surfaceCmds["surface:2"], "ls -al") {
		t.Errorf("expected surface:2 to receive ls command, got %q", surfaceCmds["surface:2"])
	}
	if !strings.Contains(surfaceCmds["surface:3"], "lazygit") {
		t.Errorf("expected surface:3 to receive lazygit, got %q", surfaceCmds["surface:3"])
	}
	if !strings.Contains(surfaceCmds["surface:1"], "cac") {
		t.Errorf("expected surface:1 to receive cac, got %q", surfaceCmds["surface:1"])
	}
}

func TestAdd_ClaudeAgentStaged(t *testing.T) {
	exec := newFake()
	exec.results["cmux list-panes"] = "* pane:1  [1 surface]  [focused]\n  pane:2  [1 surface]\n  pane:3  [1 surface]"
	exec.results["cmux list-pane-surfaces pane:1"] = "* surface:1  Terminal  [selected]"
	exec.results["cmux list-pane-surfaces pane:2"] = "* surface:2  Terminal  [selected]"
	exec.results["cmux list-pane-surfaces pane:3"] = "* surface:3  Terminal  [selected]"
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
	if text != "cac" {
		t.Errorf("expected staged cac command, got %q", text)
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
	exec.results["cmux list-panes"] = "* pane:42  [1 surface]  [focused]"
	exec.results["cmux list-pane-surfaces pane:42"] = "* surface:99  Terminal  [selected]"
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

func TestAdd_BareOKErrors(t *testing.T) {
	exec := newFake()
	exec.results["cmux workspace create"] = "OK" // no ref

	err := cmux.Add(cmux.AddOptions{CWD: "/code/myproject", Agent: "claude"}, exec)
	if err == nil {
		t.Fatal("expected error on bare 'OK' workspace create output")
	}
	for _, c := range exec.calls {
		if c.cmd == "cmux" && c.args[0] == "list-panes" {
			t.Error("downstream cmux calls should not run after parse failure")
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
