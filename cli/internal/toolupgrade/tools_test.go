package toolupgrade_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)

// fakeExec keys Run results by the full command line ("cmd arg1 arg2") so the
// same binary (e.g. brew) can return different output per argument set.
type fakeExec struct {
	calls   [][]string
	outputs map[string]string
	errs    map[string]error
	paths   map[string]bool
}

func newFakeExec() *fakeExec {
	return &fakeExec{outputs: map[string]string{}, errs: map[string]error{}, paths: map[string]bool{}}
}

func cmdKey(cmd string, args []string) string {
	return strings.TrimSpace(cmd + " " + strings.Join(args, " "))
}

func (f *fakeExec) Run(cmd string, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string{cmd}, args...))
	k := cmdKey(cmd, args)
	return f.outputs[k], f.errs[k]
}

func (f *fakeExec) LookPath(name string) (string, error) {
	if f.paths[name] {
		return "/opt/homebrew/bin/" + name, nil
	}
	return "", errors.New("not found")
}

// findTool returns the registered tool with the given name.
func findTool(t *testing.T, name string) toolupgrade.Tool {
	t.Helper()
	for _, tool := range toolupgrade.Registry() {
		if tool.Name() == name {
			return tool
		}
	}
	t.Fatalf("tool %q not in registry", name)
	return nil
}

func TestRegistryContainsExpectedTools(t *testing.T) {
	got := map[string]bool{}
	for _, tool := range toolupgrade.Registry() {
		got[tool.Name()] = true
	}
	for _, want := range []string{"brew", "claude", "opencode"} {
		if !got[want] {
			t.Errorf("registry missing %q", want)
		}
	}
}

func TestBrewVersions(t *testing.T) {
	exec := newFakeExec()
	exec.paths["brew"] = true
	exec.outputs["brew --version"] = "Homebrew 5.1.14\nHomebrew/homebrew-core (git revision abc)"
	exec.outputs["curl -fsSL https://api.github.com/repos/Homebrew/brew/releases/latest"] = `{"tag_name":"5.1.15"}`

	brew := findTool(t, "brew")
	if !brew.Installed(exec) {
		t.Error("brew should be installed")
	}
	if cur, _ := brew.CurrentVersion(exec); cur != "5.1.14" {
		t.Errorf("current: want 5.1.14, got %q", cur)
	}
	if latest, _ := brew.LatestVersion(exec); latest != "5.1.15" {
		t.Errorf("latest: want 5.1.15, got %q", latest)
	}
}

func TestClaudeVersions(t *testing.T) {
	exec := newFakeExec()
	exec.paths["claude"] = true
	exec.outputs["claude --version"] = "2.1.165 (Claude Code)"

	claude := findTool(t, "claude")
	if cur, _ := claude.CurrentVersion(exec); cur != "2.1.165" {
		t.Errorf("current: want 2.1.165, got %q", cur)
	}
	latest, err := claude.LatestVersion(exec)
	if err != nil {
		t.Fatal(err)
	}
	if latest != "" {
		t.Errorf("claude latest should be unknown (\"\"), got %q", latest)
	}
}

func TestOpencodeVersionsStripsVPrefix(t *testing.T) {
	exec := newFakeExec()
	exec.paths["opencode"] = true
	exec.outputs["opencode --version"] = "1.2.10"
	exec.outputs["curl -fsSL https://api.github.com/repos/sst/opencode/releases/latest"] = `{"tag_name":"v1.16.0"}`

	oc := findTool(t, "opencode")
	if cur, _ := oc.CurrentVersion(exec); cur != "1.2.10" {
		t.Errorf("current: want 1.2.10, got %q", cur)
	}
	if latest, _ := oc.LatestVersion(exec); latest != "1.16.0" {
		t.Errorf("latest: want 1.16.0 (v stripped), got %q", latest)
	}
}

func TestNotInstalledWhenLookPathFails(t *testing.T) {
	exec := newFakeExec() // no paths set
	if findTool(t, "brew").Installed(exec) {
		t.Error("brew should report not installed when LookPath fails")
	}
}

func TestUpgradeCommands(t *testing.T) {
	cases := []struct {
		tool     string
		wantArgs []string
	}{
		{"brew", []string{"brew", "update"}},
		{"claude", []string{"claude", "update"}},
		{"opencode", []string{"opencode", "upgrade"}},
	}
	for _, c := range cases {
		t.Run(c.tool, func(t *testing.T) {
			exec := newFakeExec()
			if err := findTool(t, c.tool).Upgrade(exec); err != nil {
				t.Fatal(err)
			}
			if len(exec.calls) != 1 {
				t.Fatalf("want 1 call, got %d: %v", len(exec.calls), exec.calls)
			}
			got := exec.calls[0]
			if strings.Join(got, " ") != strings.Join(c.wantArgs, " ") {
				t.Errorf("want %v, got %v", c.wantArgs, got)
			}
		})
	}
}

func TestUpgradeWrapsError(t *testing.T) {
	exec := newFakeExec()
	exec.errs["brew update"] = errors.New("network down")
	if err := findTool(t, "brew").Upgrade(exec); err == nil {
		t.Fatal("expected error when brew update fails")
	}
}
