package brew_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

	t.Run("uses temp file not hardcoded path", func(t *testing.T) {
		exec := newFake()
		if err := brew.Sync("/repo/.Brewfile", exec); err != nil {
			t.Fatal(err)
		}
		// brew dump arg must not be the old hardcoded path
		for _, c := range exec.calls {
			for _, a := range c.args {
				if a == "--file=/tmp/.Brewfile.current" {
					t.Error("used hardcoded /tmp/.Brewfile.current; expected os.CreateTemp path")
				}
			}
		}
	})

	t.Run("returns error on brew dump failure", func(t *testing.T) {
		exec := newFake()
		exec.errs["brew"] = errors.New("brew not found")
		if err := brew.Sync("/repo/.Brewfile", exec); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when diff fails with non-exit-code error", func(t *testing.T) {
		fake := &sequentialFake{results: []struct {
			out string
			err error
		}{
			{out: "", err: nil},                   // brew dump succeeds
			{out: "", err: errors.New("diff: -")}, // diff hard-fails (not an ExitError)
		}}
		if err := brew.Sync("/repo/.Brewfile", fake); err == nil {
			t.Fatal("expected error from diff failure, got nil")
		}
	})
}

func TestBrewfilePath(t *testing.T) {
	machine, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	if idx := strings.Index(machine, "."); idx >= 0 {
		machine = machine[:idx]
	}

	t.Run("returns host Brewfile when present", func(t *testing.T) {
		repo := t.TempDir()
		hostBrewfile := filepath.Join(repo, "hosts", machine, ".Brewfile")
		if err := os.MkdirAll(filepath.Dir(hostBrewfile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(hostBrewfile, []byte("brew \"git\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := brew.BrewfilePath(repo)
		if err != nil {
			t.Fatal(err)
		}
		if got != hostBrewfile {
			t.Errorf("want %s, got %s", hostBrewfile, got)
		}
	})

	t.Run("falls back to dotfiles Brewfile when no host override", func(t *testing.T) {
		repo := t.TempDir()
		dotBrewfile := filepath.Join(repo, "dotfiles", ".Brewfile")

		got, err := brew.BrewfilePath(repo)
		if err != nil {
			t.Fatal(err)
		}
		if got != dotBrewfile {
			t.Errorf("want %s, got %s", dotBrewfile, got)
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

// sequentialFake returns preset outputs/errors per call index.
type sequentialFake struct {
	results []struct {
		out string
		err error
	}
	idx int
}

func (s *sequentialFake) Run(_ string, _ ...string) (string, error) {
	if s.idx >= len(s.results) {
		return "", nil
	}
	r := s.results[s.idx]
	s.idx++
	return r.out, r.err
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

	t.Run("includes command output in error when bundle install fails", func(t *testing.T) {
		fake := &sequentialFake{results: []struct {
			out string
			err error
		}{
			{out: "", err: nil},                                             // brew update succeeds
			{out: "Error: some-pkg: not found", err: errors.New("exit 1")}, // bundle install fails
		}}
		err := brew.Install(fake)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "Error: some-pkg: not found") {
			t.Errorf("error message missing brew output: %s", msg)
		}
	})
}
