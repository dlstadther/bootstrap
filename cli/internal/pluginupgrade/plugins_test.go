package pluginupgrade_test

import (
	"fmt"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
)

const samplePluginList = `Installed plugins:

  ❯ agentsmith@dlstadther-agentsmith
    Version: 1.0.0
    Scope: user
    Status: ✔ enabled

  ❯ beads@local
    Version: 1.0.0
    Scope: user
    Status: ✘ disabled

  ❯ superpowers@claude-plugins-official
    Version: 5.1.0
    Scope: user
    Status: ✔ enabled

  ❯ code-review@claude-plugins-official
    Version: unknown
    Scope: user
    Status: ✔ enabled

  ❯ context7@claude-plugins-official
    Version: unknown
    Scope: user
    Status: ✘ disabled`

func TestParsePluginList_OnlyEnabled(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	if len(tools) != 3 {
		t.Fatalf("expected 3 enabled plugins, got %d", len(tools))
	}
}

func TestParsePluginList_Names(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name()] = true
	}
	want := []string{
		"agentsmith@dlstadther-agentsmith",
		"superpowers@claude-plugins-official",
		"code-review@claude-plugins-official",
	}
	for _, w := range want {
		if !names[w] {
			t.Errorf("missing plugin %q", w)
		}
	}
	if names["beads@local"] {
		t.Error("disabled plugin beads@local must not be included")
	}
	if names["context7@claude-plugins-official"] {
		t.Error("disabled plugin context7@claude-plugins-official must not be included")
	}
}

func TestParsePluginList_KnownVersion(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	for _, tool := range tools {
		if tool.Name() == "agentsmith@dlstadther-agentsmith" {
			ver, err := tool.CurrentVersion(nil)
			if err != nil {
				t.Fatal(err)
			}
			if ver != "1.0.0" {
				t.Fatalf("expected 1.0.0, got %q", ver)
			}
			return
		}
	}
	t.Fatal("agentsmith plugin not found")
}

func TestParsePluginList_UnknownVersion_ReturnsEmpty(t *testing.T) {
	tools := pluginupgrade.ParsePluginList(samplePluginList)
	for _, tool := range tools {
		if tool.Name() == "code-review@claude-plugins-official" {
			ver, err := tool.CurrentVersion(nil)
			if err != nil {
				t.Fatal(err)
			}
			if ver != "" {
				t.Fatalf("expected empty string for unknown version, got %q", ver)
			}
			return
		}
	}
	t.Fatal("code-review plugin not found")
}

func TestParsePluginList_Empty(t *testing.T) {
	tools := pluginupgrade.ParsePluginList("")
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestDiscover_RunsClaudePluginsList(t *testing.T) {
	e := &stubExec{
		outputs: map[string]string{
			"claude plugins list": samplePluginList,
		},
		errs: map[string]error{},
	}
	tools, err := pluginupgrade.Discover(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	called := false
	for _, c := range e.called {
		if c == "claude plugins list" {
			called = true
		}
	}
	if !called {
		t.Fatal("expected 'claude plugins list' to be called")
	}
}

func TestDiscover_ErrorPropagates(t *testing.T) {
	e := &stubExec{
		outputs: map[string]string{},
		errs:    map[string]error{"claude plugins list": fmt.Errorf("fake error")},
	}
	_, err := pluginupgrade.Discover(e)
	if err == nil {
		t.Fatal("expected error from Discover when exec fails")
	}
}
