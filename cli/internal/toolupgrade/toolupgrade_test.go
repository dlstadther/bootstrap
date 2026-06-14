package toolupgrade_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)

// fakeTool is a controllable Tool for orchestration tests.
type fakeTool struct {
	name       string
	installed  bool
	current    string
	latest     string
	currentErr error
	latestErr  error
	upgradeErr error
	upgraded   bool
}

func (t *fakeTool) Name() string                          { return t.name }
func (t *fakeTool) Installed(iexec.LookPathExecutor) bool { return t.installed }
func (t *fakeTool) CurrentVersion(iexec.LookPathExecutor) (string, error) {
	return t.current, t.currentErr
}
func (t *fakeTool) LatestVersion(iexec.LookPathExecutor) (string, error) {
	return t.latest, t.latestErr
}
func (t *fakeTool) Upgrade(iexec.LookPathExecutor) error {
	if t.upgradeErr != nil {
		return t.upgradeErr
	}
	t.upgraded = true
	return nil
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name string
		tool *fakeTool
		want toolupgrade.State
	}{
		{"not installed", &fakeTool{name: "x", installed: false}, toolupgrade.StateNotInstalled},
		{"up to date", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: "1.0.0"}, toolupgrade.StateUpToDate},
		{"update available", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: "1.1.0"}, toolupgrade.StateUpdateAvailable},
		{"latest unknown", &fakeTool{name: "x", installed: true, current: "1.0.0", latest: ""}, toolupgrade.StateUnknown},
		{"current version error", &fakeTool{name: "x", installed: true, currentErr: errors.New("cmd failed")}, toolupgrade.StateUnknown},
		{"latest version error", &fakeTool{name: "x", installed: true, current: "1.0.0", latestErr: errors.New("cmd failed")}, toolupgrade.StateUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := toolupgrade.Evaluate(c.tool, nil)
			if got.State != c.want {
				t.Errorf("state: want %v, got %v", c.want, got.State)
			}
		})
	}
}

func TestRunCheckModeRunsNothing(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0"}
	deciderCalled := false
	decider := func([]toolupgrade.Status) (map[string]bool, error) {
		deciderCalled = true
		return map[string]bool{"a": true}, nil
	}
	var buf bytes.Buffer
	err := toolupgrade.Run(
		toolupgrade.Options{Check: true, Out: &buf},
		nil,
		[]toolupgrade.Tool{a},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if deciderCalled {
		t.Error("decider must not be called in --check mode")
	}
	if a.upgraded {
		t.Error("no upgrade may run in --check mode")
	}
	if !strings.Contains(buf.String(), "update available") {
		t.Errorf("table missing status; got: %q", buf.String())
	}
}

func TestRunOnlyUpgradesApprovedAfterPrompting(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0"}
	b := &fakeTool{name: "b", installed: true, current: "2.0.0", latest: "2.5.0"}

	var sawCandidates []string
	decider := func(cands []toolupgrade.Status) (map[string]bool, error) {
		for _, c := range cands {
			sawCandidates = append(sawCandidates, c.Name)
		}
		if a.upgraded || b.upgraded {
			t.Fatal("an upgrade ran before the decider returned")
		}
		return map[string]bool{"a": true, "b": false}, nil
	}
	var buf bytes.Buffer
	err := toolupgrade.Run(
		toolupgrade.Options{Out: &buf},
		nil,
		[]toolupgrade.Tool{a, b},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(sawCandidates) != 2 {
		t.Fatalf("decider should see 2 candidates, saw %v", sawCandidates)
	}
	if !a.upgraded {
		t.Error("approved tool a was not upgraded")
	}
	if b.upgraded {
		t.Error("unapproved tool b was upgraded")
	}
	if !strings.Contains(buf.String(), "1 upgraded, 1 skipped, 0 failed") {
		t.Errorf("summary wrong; got: %q", buf.String())
	}
}

func TestRunContinuesAfterFailure(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.1.0", upgradeErr: errors.New("boom")}
	b := &fakeTool{name: "b", installed: true, current: "2.0.0", latest: "2.5.0"}
	decider := func([]toolupgrade.Status) (map[string]bool, error) {
		return map[string]bool{"a": true, "b": true}, nil
	}
	var buf bytes.Buffer
	if err := toolupgrade.Run(toolupgrade.Options{Out: &buf}, nil, []toolupgrade.Tool{a, b}, decider); err != nil {
		t.Fatal(err)
	}
	if !b.upgraded {
		t.Error("tool b should still upgrade after a fails")
	}
	if !strings.Contains(buf.String(), "1 upgraded, 0 skipped, 1 failed") {
		t.Errorf("summary wrong; got: %q", buf.String())
	}
}

type failWriter struct{ err error }

func (f failWriter) Write([]byte) (int, error) { return 0, f.err }

func TestRunReturnsWriteError(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.0.0"}
	decider := func([]toolupgrade.Status) (map[string]bool, error) { return nil, nil }
	err := toolupgrade.Run(toolupgrade.Options{Out: failWriter{err: errors.New("disk full")}}, nil, []toolupgrade.Tool{a}, decider)
	if err == nil {
		t.Fatal("expected write error to surface")
	}
}

func TestRunNoCandidatesSkipsDecider(t *testing.T) {
	a := &fakeTool{name: "a", installed: true, current: "1.0.0", latest: "1.0.0"}
	called := false
	decider := func([]toolupgrade.Status) (map[string]bool, error) { called = true; return nil, nil }
	var buf bytes.Buffer
	if err := toolupgrade.Run(toolupgrade.Options{Out: &buf}, nil, []toolupgrade.Tool{a}, decider); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("decider must not be called when nothing is upgradable")
	}
	if !strings.Contains(buf.String(), "All tools up to date") {
		t.Errorf("expected up-to-date message; got: %q", buf.String())
	}
}

func TestStdinDeciderCollectsAllAnswers(t *testing.T) {
	cands := []toolupgrade.Status{
		{Name: "a", Current: "1.0.0", Latest: "1.1.0"},
		{Name: "b", Current: "2.0.0", Latest: ""},
	}
	in := strings.NewReader("y\nn\n")
	var out bytes.Buffer
	decider := toolupgrade.StdinDecider(in, &out)
	approved, err := decider(cands)
	if err != nil {
		t.Fatal(err)
	}
	if !approved["a"] || approved["b"] {
		t.Errorf("approvals wrong: %v", approved)
	}
	if !strings.Contains(out.String(), "Upgrade b (2.0.0 → ?)") {
		t.Errorf("unknown-latest prompt should show '?'; got: %q", out.String())
	}
}
