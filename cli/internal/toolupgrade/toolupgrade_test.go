package toolupgrade_test

import (
	"testing"

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

func (t *fakeTool) Name() string                                          { return t.name }
func (t *fakeTool) Installed(toolupgrade.Executor) bool                   { return t.installed }
func (t *fakeTool) CurrentVersion(toolupgrade.Executor) (string, error)   { return t.current, t.currentErr }
func (t *fakeTool) LatestVersion(toolupgrade.Executor) (string, error)    { return t.latest, t.latestErr }
func (t *fakeTool) Upgrade(toolupgrade.Executor) error {
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
