package cmd

import (
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/version"
)

// nopExec satisfies iexec.Executor; runAudit's error path returns before it.
type nopExec struct{}

func (nopExec) Run(_ string, _ ...string) (string, error) { return "", nil }

func TestRunAudit_NoRepoPath(t *testing.T) {
	orig := version.RepoPath
	version.RepoPath = ""
	defer func() { version.RepoPath = orig }()

	getenv := func(string) string { return "" }
	err := runAudit(getenv, nopExec{}, false)
	if err == nil {
		t.Fatal("expected error when repo path is unresolved")
	}
	if !strings.Contains(err.Error(), "repo path is not set") {
		t.Errorf("unexpected error: %v", err)
	}
}
