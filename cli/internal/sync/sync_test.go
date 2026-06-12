package sync_test

import (
	"errors"
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
