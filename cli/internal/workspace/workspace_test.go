package workspace_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/workspace"
)

func TestValidateAgent(t *testing.T) {
	for _, a := range workspace.AllowedAgents {
		if err := workspace.ValidateAgent(a); err != nil {
			t.Errorf("agent %q should be valid, got %v", a, err)
		}
	}
	err := workspace.ValidateAgent("nope")
	if err == nil || !strings.Contains(err.Error(), "invalid agent") {
		t.Fatalf("expected invalid agent error, got %v", err)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}

	got, err := workspace.ExpandHome("~/code")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "code"); got != want {
		t.Errorf("expand ~/code: want %q, got %q", want, got)
	}

	// Paths without a leading ~/ are returned unchanged.
	for _, p := range []string{"/abs/path", "relative", "~user/notexpanded"} {
		got, err := workspace.ExpandHome(p)
		if err != nil {
			t.Fatal(err)
		}
		if got != p {
			t.Errorf("expand %q: want unchanged, got %q", p, got)
		}
	}
}

func TestCoalesce(t *testing.T) {
	if got := workspace.Coalesce("a", "b"); got != "a" {
		t.Errorf("want a, got %q", got)
	}
	if got := workspace.Coalesce("", "b"); got != "b" {
		t.Errorf("want b, got %q", got)
	}
	if got := workspace.Coalesce("", ""); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

// parseLines splits raw file contents into trimmed non-empty lines — a trivial
// parser used only to exercise LoadDir.
func parseLines(data []byte, _ string) (int, error) {
	return len(strings.Fields(string(data))), nil
}

func TestLoadDir_ParsesMatchingExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one two"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("three"), 0o644)
	os.WriteFile(filepath.Join(dir, "skip.md"), []byte("ignored"), 0o644)

	got, err := workspace.LoadDir(dir, ".txt", parseLines)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 parsed files, got %d (%v)", len(got), got)
	}
}

func TestLoadDir_MissingDir(t *testing.T) {
	got, err := workspace.LoadDir("/nonexistent/dir", ".txt", parseLines)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("missing dir: want nil, got %v", got)
	}
}

func TestLoadDir_ParseErrorAborts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.txt"), []byte("boom"), 0o644)

	sentinel := errors.New("parse failed")
	_, err := workspace.LoadDir(dir, ".txt", func([]byte, string) (int, error) {
		return 0, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected parse error to propagate, got %v", err)
	}
}

// fakeBackend records the order phases run and can fail any phase.
type fakeBackend struct {
	phases   []string
	failOn   string
	failErr  error
	overrode bool
}

func (b *fakeBackend) run(phase string) error {
	b.phases = append(b.phases, phase)
	if b.failOn == phase {
		return b.failErr
	}
	return nil
}

func (b *fakeBackend) EnsureRunning() error { return b.run("ensure") }
func (b *fakeBackend) Restore() error       { return b.run("restore") }
func (b *fakeBackend) Override()            { b.overrode = true; _ = b.run("override") }
func (b *fakeBackend) Create() error        { return b.run("create") }
func (b *fakeBackend) TearDown() error      { return b.run("teardown") }

func TestStart_PhaseOrder(t *testing.T) {
	b := &fakeBackend{}
	if err := workspace.Start(workspace.StartOptions{Override: true}, b); err != nil {
		t.Fatal(err)
	}
	want := []string{"ensure", "restore", "override", "create"}
	if strings.Join(b.phases, ",") != strings.Join(want, ",") {
		t.Errorf("phase order: want %v, got %v", want, b.phases)
	}
}

func TestStart_NoRestoreSkipsRestore(t *testing.T) {
	b := &fakeBackend{}
	if err := workspace.Start(workspace.StartOptions{NoRestore: true}, b); err != nil {
		t.Fatal(err)
	}
	for _, p := range b.phases {
		if p == "restore" {
			t.Error("restore should be skipped with NoRestore")
		}
	}
}

func TestStart_NoOverrideSkipsOverride(t *testing.T) {
	b := &fakeBackend{}
	if err := workspace.Start(workspace.StartOptions{NoRestore: true}, b); err != nil {
		t.Fatal(err)
	}
	if b.overrode {
		t.Error("override should not run without Override")
	}
}

func TestStart_EnsureFailureAborts(t *testing.T) {
	sentinel := errors.New("down")
	b := &fakeBackend{failOn: "ensure", failErr: sentinel}
	err := workspace.Start(workspace.StartOptions{}, b)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected ensure error, got %v", err)
	}
	if len(b.phases) != 1 {
		t.Errorf("expected to stop after ensure, ran %v", b.phases)
	}
}

func TestReset_TearsDownThenStartsWithoutRestore(t *testing.T) {
	b := &fakeBackend{}
	if err := workspace.Reset(b); err != nil {
		t.Fatal(err)
	}
	want := []string{"teardown", "ensure", "create"}
	if strings.Join(b.phases, ",") != strings.Join(want, ",") {
		t.Errorf("reset phases: want %v, got %v", want, b.phases)
	}
}

func TestReset_TearDownFailureAborts(t *testing.T) {
	sentinel := errors.New("ping failed")
	b := &fakeBackend{failOn: "teardown", failErr: sentinel}
	err := workspace.Reset(b)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected teardown error, got %v", err)
	}
	for _, p := range b.phases {
		if p != "teardown" {
			t.Errorf("no phase should run after teardown failure, got %v", b.phases)
		}
	}
}
