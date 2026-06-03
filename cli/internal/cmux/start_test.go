package cmux_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/cmux"
)

func TestLoadWorkspaces_Empty(t *testing.T) {
	ws, err := cmux.LoadWorkspaces(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(ws))
	}
}

func TestLoadWorkspaces_MissingDir(t *testing.T) {
	ws, err := cmux.LoadWorkspaces("/nonexistent/dir")
	if err != nil {
		t.Fatal(err)
	}
	if ws != nil {
		t.Errorf("expected nil, got %v", ws)
	}
}

func TestLoadWorkspaces_Parses(t *testing.T) {
	dir := t.TempDir()
	wc := cmux.WorkspaceConfig{
		Name: "myproject",
		CWD:  "~/code/myproject",
		Panes: []cmux.PaneSpec{
			{Command: "claude agents", NoEnter: true},
			{Split: "right", Command: "ls -al"},
		},
	}
	data, _ := json.Marshal(wc)
	os.WriteFile(filepath.Join(dir, "myproject.json"), data, 0644)

	ws, err := cmux.LoadWorkspaces(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ws))
	}
	if ws[0].Name != "myproject" {
		t.Errorf("expected name myproject, got %s", ws[0].Name)
	}
	if len(ws[0].Panes) != 2 {
		t.Errorf("expected 2 panes, got %d", len(ws[0].Panes))
	}
	if !ws[0].Panes[0].NoEnter {
		t.Error("expected first pane NoEnter=true")
	}
	if ws[0].Panes[1].Split != "right" {
		t.Errorf("expected second pane split=right, got %s", ws[0].Panes[1].Split)
	}
}

func TestLoadWorkspaces_SkipsNonJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0644)
	os.WriteFile(filepath.Join(dir, "ws.json"), []byte(`{"name":"x","cwd":"/tmp"}`), 0644)

	ws, err := cmux.LoadWorkspaces(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 {
		t.Errorf("expected 1, got %d", len(ws))
	}
}

func TestStart_CmuxNotRunning(t *testing.T) {
	f := newFake()
	f.errs["cmux ping"] = errors.New("connection refused")
	err := cmux.Start(cmux.StartOptions{NoRestore: true, WorkspacesDir: t.TempDir()}, f)
	if err == nil || !strings.Contains(err.Error(), "cmux is not running") {
		t.Fatalf("expected cmux not running error, got %v", err)
	}
}

func TestStart_NoRestore(t *testing.T) {
	f := newFake()
	err := cmux.Start(cmux.StartOptions{NoRestore: true, WorkspacesDir: t.TempDir()}, f)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "restore-session" {
			t.Error("restore-session should not be called with --no-restore")
		}
	}
}

func TestStart_Restore(t *testing.T) {
	f := newFake()
	err := cmux.Start(cmux.StartOptions{NoRestore: false, WorkspacesDir: t.TempDir()}, f)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "restore-session" {
			found = true
		}
	}
	if !found {
		t.Error("expected restore-session call")
	}
}

func TestStart_CreatesWorkspace(t *testing.T) {
	f := newFake()
	dir := t.TempDir()
	wc := cmux.WorkspaceConfig{
		Name: "myproject",
		CWD:  "/code/myproject",
		Panes: []cmux.PaneSpec{
			{Command: "claude agents", NoEnter: true},
			{Split: "right", Command: "ls -al"},
		},
	}
	data, _ := json.Marshal(wc)
	os.WriteFile(filepath.Join(dir, "myproject.json"), data, 0644)

	if err := cmux.Start(cmux.StartOptions{NoRestore: true, WorkspacesDir: dir}, f); err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-workspace" {
			found = true
			nameIdx := indexOf(c.args, "--name")
			if nameIdx >= 0 && c.args[nameIdx+1] != "myproject" {
				t.Errorf("expected --name myproject, got %s", c.args[nameIdx+1])
			}
		}
	}
	if !found {
		t.Error("expected new-workspace call")
	}
}

func TestStart_SkipsExistingWorkspace(t *testing.T) {
	f := newFake()
	f.results["cmux list-workspaces"] = "workspace:1 myproject"

	dir := t.TempDir()
	wc := cmux.WorkspaceConfig{Name: "myproject", CWD: "/code/myproject"}
	data, _ := json.Marshal(wc)
	os.WriteFile(filepath.Join(dir, "myproject.json"), data, 0644)

	if err := cmux.Start(cmux.StartOptions{NoRestore: true, WorkspacesDir: dir}, f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-workspace" {
			t.Error("new-workspace should not be called for existing workspace")
		}
	}
}

func TestStart_Override(t *testing.T) {
	f := newFake()
	f.results["cmux list-workspaces"] = "workspace:1 myproject"

	dir := t.TempDir()
	wc := cmux.WorkspaceConfig{Name: "myproject", CWD: "/code/myproject"}
	data, _ := json.Marshal(wc)
	os.WriteFile(filepath.Join(dir, "myproject.json"), data, 0644)

	if err := cmux.Start(cmux.StartOptions{NoRestore: true, Override: true, WorkspacesDir: dir}, f); err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "close-workspace" {
			wsIdx := indexOf(c.args, "--workspace")
			if wsIdx >= 0 && c.args[wsIdx+1] == "workspace:1" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected close-workspace call targeting workspace:1")
	}
}

func TestStart_SplitsCreatedInOrder(t *testing.T) {
	f := newFake()
	dir := t.TempDir()
	wc := cmux.WorkspaceConfig{
		Name: "myproject",
		CWD:  "/code/myproject",
		Panes: []cmux.PaneSpec{
			{Command: "agent", NoEnter: true},
			{Split: "right", Command: "ls"},
			{Split: "down", Command: "lazygit"},
		},
	}
	data, _ := json.Marshal(wc)
	os.WriteFile(filepath.Join(dir, "myproject.json"), data, 0644)

	if err := cmux.Start(cmux.StartOptions{NoRestore: true, WorkspacesDir: dir}, f); err != nil {
		t.Fatal(err)
	}

	var splits []string
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-split" {
			splits = append(splits, c.args[1])
		}
	}
	if len(splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(splits))
	}
	if splits[0] != "right" || splits[1] != "down" {
		t.Errorf("expected right+down, got %v", splits)
	}
}

func TestStart_LocalWorkspacesDir(t *testing.T) {
	f := newFake()

	mainDir := t.TempDir()
	localDir := t.TempDir()

	writeWS := func(dir, name string) {
		wc := cmux.WorkspaceConfig{Name: name, CWD: "/code/" + name}
		data, _ := json.Marshal(wc)
		os.WriteFile(filepath.Join(dir, name+".json"), data, 0644)
	}
	writeWS(mainDir, "proj-a")
	writeWS(localDir, "proj-b")

	if err := cmux.Start(cmux.StartOptions{
		NoRestore:          true,
		WorkspacesDir:      mainDir,
		LocalWorkspacesDir: localDir,
	}, f); err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 0 && c.args[0] == "new-workspace" {
			nameIdx := indexOf(c.args, "--name")
			if nameIdx >= 0 {
				names[c.args[nameIdx+1]] = true
			}
		}
	}
	if !names["proj-a"] || !names["proj-b"] {
		t.Errorf("expected both proj-a and proj-b created, got %v", names)
	}
}
