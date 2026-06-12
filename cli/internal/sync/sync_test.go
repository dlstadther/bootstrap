package sync_test

import (
	"errors"
	"os"
	"testing"

	isync "github.com/dlstadther/bootstrap/cli/internal/sync"
)

type call struct {
	cmd  string
	args []string
}

type response struct {
	output string
	err    error
}

type fakeExec struct {
	calls     []call
	responses []response
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, call{cmd: cmd, args: args})
	if len(f.responses) == 0 {
		return "", nil
	}
	r := f.responses[0]
	f.responses = f.responses[1:]
	return r.output, r.err
}

func TestSyncMise(t *testing.T) {
	t.Run("calls mise install", func(t *testing.T) {
		exec := &fakeExec{}
		if err := isync.SyncMise(exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 1 || exec.calls[0].cmd != "mise" {
			t.Fatalf("expected 1 mise call, got %v", exec.calls)
		}
		if exec.calls[0].args[0] != "install" {
			t.Errorf("expected mise install, got args %v", exec.calls[0].args)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		exec := &fakeExec{responses: []response{{err: errors.New("mise not found")}}}
		if err := isync.SyncMise(exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestSyncBrew(t *testing.T) {
	t.Run("check passes, skips install", func(t *testing.T) {
		exec := &fakeExec{} // no error = check exits 0
		if err := isync.SyncBrew(exec, false); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(exec.calls))
		}
		if exec.calls[0].args[1] != "check" {
			t.Errorf("expected bundle check, got %v", exec.calls[0].args)
		}
	})

	t.Run("check fails, runs install", func(t *testing.T) {
		exec := &fakeExec{responses: []response{
			{err: errors.New("missing packages")}, // check fails
			{},                                     // install succeeds
		}}
		if err := isync.SyncBrew(exec, false); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(exec.calls))
		}
		if exec.calls[1].args[1] != "install" {
			t.Errorf("expected bundle install, got %v", exec.calls[1].args)
		}
	})

	t.Run("force skips check, runs install", func(t *testing.T) {
		exec := &fakeExec{}
		if err := isync.SyncBrew(exec, true); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(exec.calls))
		}
		if exec.calls[0].args[1] != "install" {
			t.Errorf("expected bundle install, got %v", exec.calls[0].args)
		}
	})

	t.Run("install failure returns error", func(t *testing.T) {
		exec := &fakeExec{responses: []response{
			{err: errors.New("missing packages")}, // check fails
			{err: errors.New("network error")},     // install fails
		}}
		if err := isync.SyncBrew(exec, false); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func writeSettings(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "settings*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestSyncPlugins(t *testing.T) {
	t.Run("installs only enabled plugins in sorted order", func(t *testing.T) {
		path := writeSettings(t, `{"enabledPlugins":{"plugin-a":true,"plugin-b":false,"plugin-c":true}}`)
		exec := &fakeExec{}
		if err := isync.SyncPlugins(path, exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 2 {
			t.Fatalf("expected 2 calls, got %d: %v", len(exec.calls), exec.calls)
		}
		lastName := func(c call) string { return c.args[len(c.args)-1] }
		if lastName(exec.calls[0]) != "plugin-a" {
			t.Errorf("first install: want plugin-a, got %v", exec.calls[0].args)
		}
		if lastName(exec.calls[1]) != "plugin-c" {
			t.Errorf("second install: want plugin-c, got %v", exec.calls[1].args)
		}
	})

	t.Run("continues on individual failure and returns combined error", func(t *testing.T) {
		path := writeSettings(t, `{"enabledPlugins":{"plugin-a":true,"plugin-b":true}}`)
		exec := &fakeExec{responses: []response{
			{err: errors.New("install failed")},
			{},
		}}
		if err := isync.SyncPlugins(path, exec); err == nil {
			t.Fatal("expected error, got nil")
		}
		if len(exec.calls) != 2 {
			t.Fatalf("expected both plugins attempted, got %d calls", len(exec.calls))
		}
	})

	t.Run("returns error when settings file missing", func(t *testing.T) {
		exec := &fakeExec{}
		if err := isync.SyncPlugins("/nonexistent/settings.json", exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("no-op when no plugins enabled", func(t *testing.T) {
		path := writeSettings(t, `{"enabledPlugins":{"plugin-a":false}}`)
		exec := &fakeExec{}
		if err := isync.SyncPlugins(path, exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 0 {
			t.Fatalf("expected 0 calls, got %d", len(exec.calls))
		}
	})
}
