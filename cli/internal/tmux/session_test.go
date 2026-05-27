package tmux_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/tmux"
)

var errFake = errors.New("fake error")

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

func TestParseSession_Name(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "bootstrap" {
		t.Errorf("name: want bootstrap, got %q", s.Name)
	}
}

func TestParseSession_Root(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if s.Root != "~/code/bootstrap" {
		t.Errorf("root: want ~/code/bootstrap, got %q", s.Root)
	}
}

func TestParseSession_WindowName(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Windows) != 1 {
		t.Fatalf("windows: want 1, got %d", len(s.Windows))
	}
	if s.Windows[0].Name != "main" {
		t.Errorf("window name: want main, got %q", s.Windows[0].Name)
	}
}

func TestParseSession_WindowLayout(t *testing.T) {
	s, err := tmux.ParseSession([]byte(basicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if s.Windows[0].Layout != "main-vertical" {
		t.Errorf("layout: want main-vertical, got %q", s.Windows[0].Layout)
	}
}

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

const singlePaneYAML = `
name: admin
windows:
  - agentsview: agentsview update && agentsview serve
`

func TestParseSession_SinglePaneWindow(t *testing.T) {
	s, err := tmux.ParseSession([]byte(singlePaneYAML))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Windows) != 1 {
		t.Fatalf("windows: want 1, got %d", len(s.Windows))
	}
	w := s.Windows[0]
	if w.Name != "agentsview" {
		t.Errorf("window name: want agentsview, got %q", w.Name)
	}
	if len(w.Panes) != 1 {
		t.Fatalf("panes: want 1, got %d", len(w.Panes))
	}
	if w.Panes[0].Command != "agentsview update && agentsview serve" {
		t.Errorf("pane command: got %q", w.Panes[0].Command)
	}
}

const multiWindowYAML = `
name: admin
root: ~
windows:
  - agentsview: agentsview update && agentsview serve
  - middleman:
      root: ~/code/middleman
      panes:
        - git pull origin main && make install && middleman
`

func TestParseSession_MultipleWindows(t *testing.T) {
	s, err := tmux.ParseSession([]byte(multiWindowYAML))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Windows) != 2 {
		t.Fatalf("windows: want 2, got %d", len(s.Windows))
	}
	if s.Windows[0].Name != "agentsview" {
		t.Errorf("window[0] name: got %q", s.Windows[0].Name)
	}
	if s.Windows[1].Name != "middleman" {
		t.Errorf("window[1] name: got %q", s.Windows[1].Name)
	}
	if s.Windows[1].Root != "~/code/middleman" {
		t.Errorf("window[1] root: got %q", s.Windows[1].Root)
	}
}

func TestLoadSessions_ReadsAllYAML(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"admin.yaml":     "name: admin\nwindows:\n  - shell: bash\n",
		"bootstrap.yaml": "name: bootstrap\nwindows:\n  - main: zsh\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Non-YAML file should be ignored
	os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("ignored"), 0o644)

	sessions, err := tmux.LoadSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions: want 2, got %d", len(sessions))
	}
}

func TestLoadSessions_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	sessions, err := tmux.LoadSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions: want 0, got %d", len(sessions))
	}
}

func TestStart_NoRestore_SkipsResurrect(t *testing.T) {
	exec := newFake()
	// tmux info succeeds (tmux running)

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.yaml"), []byte("name: main\nwindows:\n  - shell: bash\n"), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	// run-shell (resurrect) should NOT appear
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 && c.args[0] == "run-shell" {
			t.Error("expected run-shell (resurrect) to be skipped with --no-restore")
		}
	}
}

func TestStart_Override_KillsMatchingSessions(t *testing.T) {
	exec := newFake()
	// existing "main" session exists
	exec.results["tmux has-session"] = ""

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.yaml"), []byte("name: main\nwindows:\n  - shell: bash\n"), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		Override:    true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	killed := false
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) >= 2 && c.args[0] == "kill-session" && c.args[1] == "-t" {
			if len(c.args) >= 3 && c.args[2] == "main" {
				killed = true
			}
		}
	}
	if !killed {
		t.Error("expected kill-session for 'main' session with --override")
	}
}

func TestStart_Override_OnlyKillsNamedSessions(t *testing.T) {
	exec := newFake()
	// all sessions exist
	exec.results["tmux has-session"] = ""

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "admin.yaml"), []byte("name: admin\nwindows:\n  - shell: bash\n"), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		Override:    true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) >= 3 && c.args[0] == "kill-session" && c.args[2] == "other" {
			t.Error("should not kill sessions not in YAML configs")
		}
	}
}

func TestStart_CreatesSessions(t *testing.T) {
	exec := newFake()
	// no existing sessions
	exec.errs["tmux has-session"] = errFake

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "work.yaml"), []byte("name: work\nwindows:\n  - main: zsh\n"), 0o644)

	err := tmux.Start(tmux.StartOptions{
		NoRestore:   true,
		SessionsDir: dir,
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	created := false
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 && c.args[0] == "new-session" {
			created = true
		}
	}
	if !created {
		t.Error("expected new-session to be called")
	}
}

func TestStart_WithRestore_InvokesResurrect(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "work.yaml"), []byte("name: work\nwindows:\n  - main: zsh\n"), 0o644)

	// create a fake resurrect script
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "restore.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho restored\n"), 0o755)

	exec := newFake()
	exec.errs["tmux has-session"] = errFake

	err := tmux.Start(tmux.StartOptions{
		NoRestore:      false,
		SessionsDir:    dir,
		ResurrectPath:  scriptPath,
		AfterRestoreWait: 0, // no sleep in tests
	}, exec)
	if err != nil {
		t.Fatal(err)
	}

	invoked := false
	for _, c := range exec.calls {
		if c.cmd == "tmux" && len(c.args) > 0 && c.args[0] == "run-shell" {
			invoked = true
		}
	}
	if !invoked {
		t.Error("expected run-shell (resurrect) to be called")
	}
}

func TestStart_WithRestore_ResurrectMissing_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "work.yaml"), []byte("name: work\nwindows:\n  - main: zsh\n"), 0o644)

	exec := newFake()
	err := tmux.Start(tmux.StartOptions{
		NoRestore:     false,
		SessionsDir:   dir,
		ResurrectPath: "/nonexistent/restore.sh",
	}, exec)
	if err == nil {
		t.Error("expected error when resurrect script is missing")
	}
}
