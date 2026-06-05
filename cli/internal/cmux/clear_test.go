package cmux_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/cmux"
)

func TestListOtherWorkspaces_CmuxNotRunning(t *testing.T) {
	f := newFake()
	f.errs["cmux ping"] = errors.New("connection refused")
	_, err := cmux.ListOtherWorkspaces("", f)
	if err == nil || !strings.Contains(err.Error(), "cmux is not running") {
		t.Fatalf("expected cmux not running error, got %v", err)
	}
}

func TestListOtherWorkspaces_ReturnsAll(t *testing.T) {
	f := newFake()
	f.results["cmux workspace list"] = `{"workspaces":[{"ref":"ws:1","title":"alpha"},{"ref":"ws:2","title":"beta"}]}`

	ws, err := cmux.ListOtherWorkspaces("", f)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(ws))
	}
}

func TestListOtherWorkspaces_SkipsCurrentRef(t *testing.T) {
	f := newFake()
	f.results["cmux workspace list"] = `{"workspaces":[{"ref":"ws:1","title":"alpha"},{"ref":"ws:2","title":"beta"}]}`

	ws, err := cmux.ListOtherWorkspaces("ws:1", f)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ws))
	}
	if ws[0].Ref != "ws:2" {
		t.Errorf("expected ws:2, got %s", ws[0].Ref)
	}
}

func TestListOtherWorkspaces_Empty(t *testing.T) {
	f := newFake()
	f.results["cmux workspace list"] = `{"workspaces":[]}`

	ws, err := cmux.ListOtherWorkspaces("", f)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(ws))
	}
}

func TestClearWorkspaces_ClosesAll(t *testing.T) {
	f := newFake()
	toClose := []cmux.WorkspaceInfo{
		{Ref: "ws:1", Title: "alpha"},
		{Ref: "ws:2", Title: "beta"},
	}

	cmux.ClearWorkspaces(toClose, f)

	closed := map[string]bool{}
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 1 && c.args[0] == "workspace" && c.args[1] == "close" {
			if len(c.args) > 2 {
				closed[c.args[2]] = true
			}
		}
	}
	if !closed["ws:1"] || !closed["ws:2"] {
		t.Errorf("expected both ws:1 and ws:2 closed, got %v", closed)
	}
}

func TestClearWorkspaces_Empty(t *testing.T) {
	f := newFake()
	cmux.ClearWorkspaces(nil, f)
	for _, c := range f.calls {
		if c.cmd == "cmux" && len(c.args) > 1 && c.args[0] == "workspace" && c.args[1] == "close" {
			t.Error("unexpected workspace close call on empty list")
		}
	}
}
