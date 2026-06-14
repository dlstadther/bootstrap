// Package paths centralizes the filesystem locations the bs CLI reads and
// writes, so command packages don't each re-derive ~/.config/... by hand.
//
// Load resolves every path env-first then falls back to home-dir defaults,
// giving callers (and tests) a single override seam.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths holds the resolved locations used by the bs CLI.
type Paths struct {
	Home string

	// TmuxSessionsDir is where `bs tmux start`/`reset` read YAML session configs.
	TmuxSessionsDir string
	// TmuxLocalSessionsDir holds machine-specific session overrides, loaded after
	// TmuxSessionsDir.
	TmuxLocalSessionsDir string

	// CmuxWorkspacesDir is where `bs cmux start`/`reset` read JSON workspace configs.
	CmuxWorkspacesDir string
	// CmuxLocalWorkspacesDir holds machine-specific workspace overrides, loaded
	// after CmuxWorkspacesDir.
	CmuxLocalWorkspacesDir string

	// ClaudeSettings is the path to ~/.claude/settings.json.
	ClaudeSettings string
}

// Load resolves all bs CLI paths. Each location honors an env override first
// (see the BS_* vars below); otherwise it derives from the user's home dir.
// The ".local" sibling directories always derive from their base so an override
// keeps its override-local pair.
func Load() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("home dir: %w", err)
	}

	tmuxSessions := envOr("BS_TMUX_SESSIONS_DIR", filepath.Join(home, ".config", "tmux", "sessions"))
	cmuxWorkspaces := envOr("BS_CMUX_WORKSPACES_DIR", filepath.Join(home, ".config", "cmux", "workspaces"))
	claudeSettings := envOr("BS_CLAUDE_SETTINGS", filepath.Join(home, ".claude", "settings.json"))

	return Paths{
		Home:                   home,
		TmuxSessionsDir:        tmuxSessions,
		TmuxLocalSessionsDir:   tmuxSessions + ".local",
		CmuxWorkspacesDir:      cmuxWorkspaces,
		CmuxLocalWorkspacesDir: cmuxWorkspaces + ".local",
		ClaudeSettings:         claudeSettings,
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
