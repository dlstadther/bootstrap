package brew_test

import (
	"errors"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/brew"
)

type call struct {
	cmd  string
	args []string
}

type fakeExec struct {
	calls   []call
	outputs map[string]string
	errs    map[string]error
}

func newFake() *fakeExec {
	return &fakeExec{outputs: map[string]string{}, errs: map[string]error{}}
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, call{cmd: cmd, args: args})
	return f.outputs[cmd], f.errs[cmd]
}

func TestSync(t *testing.T) {
	t.Run("runs brew dump then diff", func(t *testing.T) {
		exec := newFake()
		exec.outputs["diff"] = "+ new-pkg"
		if err := brew.Sync("/repo/.Brewfile", exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(exec.calls))
		}
		if exec.calls[0].cmd != "brew" {
			t.Errorf("first call: want brew, got %s", exec.calls[0].cmd)
		}
		if exec.calls[1].cmd != "diff" {
			t.Errorf("second call: want diff, got %s", exec.calls[1].cmd)
		}
	})

	t.Run("returns error on brew dump failure", func(t *testing.T) {
		exec := newFake()
		exec.errs["brew"] = errors.New("brew not found")
		if err := brew.Sync("/repo/.Brewfile", exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDump(t *testing.T) {
	t.Run("calls brew bundle dump with correct path", func(t *testing.T) {
		exec := newFake()
		if err := brew.Dump("/repo/.Brewfile", exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(exec.calls))
		}
		found := false
		for _, a := range exec.calls[0].args {
			if a == "--file=/repo/.Brewfile" {
				found = true
			}
		}
		if !found {
			t.Errorf("--file= arg not found in call: %v", exec.calls[0].args)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		exec := newFake()
		exec.errs["brew"] = errors.New("permission denied")
		if err := brew.Dump("/repo/.Brewfile", exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestInstall(t *testing.T) {
	t.Run("runs brew update then bundle install", func(t *testing.T) {
		exec := newFake()
		if err := brew.Install(exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.calls) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(exec.calls))
		}
	})

	t.Run("returns error when brew update fails", func(t *testing.T) {
		exec := newFake()
		exec.errs["brew"] = errors.New("network error")
		if err := brew.Install(exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
