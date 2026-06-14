package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/paths"
)

func TestLoad_Defaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Clear any overrides so defaults are exercised.
	t.Setenv("BS_TMUX_SESSIONS_DIR", "")
	t.Setenv("BS_CMUX_WORKSPACES_DIR", "")
	t.Setenv("BS_CLAUDE_SETTINGS", "")

	p, err := paths.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cases := map[string]string{
		"Home":                   home,
		"TmuxSessionsDir":        filepath.Join(home, ".config", "tmux", "sessions"),
		"TmuxLocalSessionsDir":   filepath.Join(home, ".config", "tmux", "sessions.local"),
		"CmuxWorkspacesDir":      filepath.Join(home, ".config", "cmux", "workspaces"),
		"CmuxLocalWorkspacesDir": filepath.Join(home, ".config", "cmux", "workspaces.local"),
		"ClaudeSettings":         filepath.Join(home, ".claude", "settings.json"),
	}
	got := map[string]string{
		"Home":                   p.Home,
		"TmuxSessionsDir":        p.TmuxSessionsDir,
		"TmuxLocalSessionsDir":   p.TmuxLocalSessionsDir,
		"CmuxWorkspacesDir":      p.CmuxWorkspacesDir,
		"CmuxLocalWorkspacesDir": p.CmuxLocalWorkspacesDir,
		"ClaudeSettings":         p.ClaudeSettings,
	}
	for k, want := range cases {
		if got[k] != want {
			t.Errorf("%s = %q, want %q", k, got[k], want)
		}
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("BS_TMUX_SESSIONS_DIR", "/custom/tmux")
	t.Setenv("BS_CMUX_WORKSPACES_DIR", "/custom/cmux")
	t.Setenv("BS_CLAUDE_SETTINGS", "/custom/settings.json")

	p, err := paths.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if p.TmuxSessionsDir != "/custom/tmux" {
		t.Errorf("TmuxSessionsDir = %q, want /custom/tmux", p.TmuxSessionsDir)
	}
	// .local sibling derives from the override, preserving the override pair.
	if p.TmuxLocalSessionsDir != "/custom/tmux.local" {
		t.Errorf("TmuxLocalSessionsDir = %q, want /custom/tmux.local", p.TmuxLocalSessionsDir)
	}
	if p.CmuxWorkspacesDir != "/custom/cmux" {
		t.Errorf("CmuxWorkspacesDir = %q, want /custom/cmux", p.CmuxWorkspacesDir)
	}
	if p.CmuxLocalWorkspacesDir != "/custom/cmux.local" {
		t.Errorf("CmuxLocalWorkspacesDir = %q, want /custom/cmux.local", p.CmuxLocalWorkspacesDir)
	}
	if p.ClaudeSettings != "/custom/settings.json" {
		t.Errorf("ClaudeSettings = %q, want /custom/settings.json", p.ClaudeSettings)
	}
}
