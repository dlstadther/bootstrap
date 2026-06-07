package pluginupgrade_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
)

// stubExec records calls and returns configured outputs/errors.
type stubExec struct {
	outputs map[string]string
	errs    map[string]error
	called  []string
}

func (s *stubExec) Run(cmd string, args ...string) (string, error) {
	key := strings.Join(append([]string{cmd}, args...), " ")
	s.called = append(s.called, key)
	if e := s.errs[key]; e != nil {
		return "", e
	}
	return s.outputs[key], nil
}

func (s *stubExec) LookPath(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

// stubTool is a configurable pluginupgrade.Tool for tests.
type stubTool struct {
	name       string
	installed  bool
	current    string
	curErr     error
	latest     string
	latErr     error
	upgradeErr error
}

func (t stubTool) Name() string                                            { return t.name }
func (t stubTool) Installed(_ pluginupgrade.Executor) bool                 { return t.installed }
func (t stubTool) CurrentVersion(_ pluginupgrade.Executor) (string, error) { return t.current, t.curErr }
func (t stubTool) LatestVersion(_ pluginupgrade.Executor) (string, error)  { return t.latest, t.latErr }
func (t stubTool) Upgrade(_ pluginupgrade.Executor) error                  { return t.upgradeErr }

func newExec() *stubExec {
	return &stubExec{outputs: map[string]string{}, errs: map[string]error{}}
}

// --- Evaluate ---

func TestEvaluate_NotInstalled(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: false}, newExec())
	if s.State != pluginupgrade.StateNotInstalled {
		t.Fatalf("got %v, want StateNotInstalled", s.State)
	}
}

func TestEvaluate_CurrentVersionError(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, curErr: errors.New("fail")}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_EmptyCurrentVersion(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: ""}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_UnknownLatest(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: ""}, newExec())
	if s.State != pluginupgrade.StateUnknown {
		t.Fatalf("got %v, want StateUnknown", s.State)
	}
}

func TestEvaluate_UpToDate(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: "1.0.0"}, newExec())
	if s.State != pluginupgrade.StateUpToDate {
		t.Fatalf("got %v, want StateUpToDate", s.State)
	}
}

func TestEvaluate_UpdateAvailable(t *testing.T) {
	s := pluginupgrade.Evaluate(stubTool{name: "x", installed: true, current: "1.0.0", latest: "2.0.0"}, newExec())
	if s.State != pluginupgrade.StateUpdateAvailable {
		t.Fatalf("got %v, want StateUpdateAvailable", s.State)
	}
	if s.Current != "1.0.0" || s.Latest != "2.0.0" {
		t.Fatalf("unexpected versions: current=%q latest=%q", s.Current, s.Latest)
	}
}

// --- Run ---

func TestRun_CheckOnly_NoDeciderCall(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a", installed: true, current: "1.0.0"}
	calledDecider := false
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		calledDecider = true
		return nil, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Check: true, Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if calledDecider {
		t.Fatal("decider must not be called in check mode")
	}
	if !strings.Contains(out.String(), "plug-a") {
		t.Fatalf("output should contain plugin name, got: %s", out.String())
	}
}

func TestRun_AllUpToDate_NoCandidates(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a", installed: true, current: "1.0.0", latest: "1.0.0"}
	calledDecider := false
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		calledDecider = true
		return nil, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if calledDecider {
		t.Fatal("decider must not be called when nothing to upgrade")
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Fatalf("expected 'up to date' message, got: %s", out.String())
	}
}

func TestRun_ApprovedUpgrade_PrintsDone(t *testing.T) {
	var out bytes.Buffer
	e := newExec()
	e.outputs["claude plugins update plug-a@mp"] = ""
	tool := stubTool{name: "plug-a@mp", installed: true, current: "1.0.0"}
	decider := pluginupgrade.Decider(func(candidates []pluginupgrade.Status) (map[string]bool, error) {
		approved := map[string]bool{}
		for _, c := range candidates {
			approved[c.Name] = true
		}
		return approved, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		e,
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "done") {
		t.Fatalf("expected 'done' in output, got: %s", output)
	}
	if !strings.Contains(output, "1 upgraded") {
		t.Fatalf("expected '1 upgraded' in summary, got: %s", output)
	}
}

func TestRun_SkippedUpgrade_PrintsSummary(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{name: "plug-a@mp", installed: true, current: "1.0.0"}
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		return map[string]bool{}, nil // approve nothing
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1 skipped") {
		t.Fatalf("expected '1 skipped', got: %s", out.String())
	}
}

func TestRun_UpgradeFailure_PrintsFailed(t *testing.T) {
	var out bytes.Buffer
	tool := stubTool{
		name:       "plug-a@mp",
		installed:  true,
		current:    "1.0.0",
		upgradeErr: errors.New("network error"),
	}
	decider := pluginupgrade.Decider(func(_ []pluginupgrade.Status) (map[string]bool, error) {
		return map[string]bool{"plug-a@mp": true}, nil
	})
	err := pluginupgrade.Run(
		pluginupgrade.Options{Out: &out},
		newExec(),
		[]pluginupgrade.Tool{tool},
		decider,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1 failed") {
		t.Fatalf("expected '1 failed', got: %s", out.String())
	}
}
