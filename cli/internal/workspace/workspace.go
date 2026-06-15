// Package workspace holds the multiplexer-agnostic logic shared by the tmux and
// cmux backends: the allowed-agent registry, path helpers, the config-directory
// loader, and the Start/Reset orchestration skeleton.
//
// tmux and cmux differ in their mechanics (session/window/pane + send-keys vs.
// workspace/surface + JSON layouts) and config formats (YAML vs. JSON), but the
// orchestration is identical in shape: ensure the multiplexer is running,
// optionally restore, optionally close conflicting workspaces, then create each
// from config. That skeleton lives here; each backend supplies the mechanics
// via the Backend interface, keeping the per-multiplexer packages thin.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AllowedAgents is the canonical set of agents that may be staged in a
// workspace's primary pane. Both tmux and cmux re-export this list.
var AllowedAgents = []string{"claude", "codex", "gemini", "opencode", "pi"}

// ValidateAgent returns nil if agent is in AllowedAgents, or a descriptive
// error naming the allowed values otherwise.
func ValidateAgent(agent string) error {
	for _, a := range AllowedAgents {
		if a == agent {
			return nil
		}
	}
	return fmt.Errorf("invalid agent %q; allowed: %s", agent, strings.Join(AllowedAgents, ", "))
}

// ExpandHome expands a leading "~/" to the user's home directory. Paths without
// that prefix are returned unchanged. A home-lookup failure is reported as an
// error so callers can decide whether to propagate or ignore it.
func ExpandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand %q: home lookup failed: %w", path, err)
		}
		return home + path[1:], nil
	}
	return path, nil
}

// Coalesce returns a if it is non-empty, otherwise b. Used to fall back from a
// window/pane-level value to its session/workspace-level default.
func Coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// LoadDir reads every file in dir whose name ends with ext (e.g. ".yaml") and
// parses it with parse, which receives the file contents and base filename.
// A missing dir yields (nil, nil); subdirectories and non-matching files are
// skipped. parse errors abort the load.
func LoadDir[T any](dir, ext string, parse func(data []byte, filename string) (T, error)) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []T
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		v, err := parse(data, e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// StartOptions are the multiplexer-agnostic flags for Start.
type StartOptions struct {
	NoRestore bool
	Override  bool
}

// Backend supplies the per-multiplexer mechanics that Start orchestrates.
// Implementations carry their own loaded config (sessions/workspaces) and
// executor; Start only sequences the phases.
type Backend interface {
	// EnsureRunning makes sure the multiplexer is up (starting it if the
	// backend supports auto-start), returning an error if it cannot be reached.
	EnsureRunning() error
	// Restore replays the previous session/workspace layout.
	Restore() error
	// Override closes any already-running workspaces named in the config so
	// they can be recreated cleanly. Errors are best-effort and swallowed.
	Override()
	// Create builds every workspace defined in the config.
	Create() error
}

// Start runs the shared start flow: ensure running, optionally restore,
// optionally override-close conflicts, then create from config.
func Start(opts StartOptions, b Backend) error {
	if err := b.EnsureRunning(); err != nil {
		return err
	}
	if !opts.NoRestore {
		if err := b.Restore(); err != nil {
			return err
		}
	}
	if opts.Override {
		b.Override()
	}
	return b.Create()
}

// ResetBackend extends Backend with the teardown step used by Reset.
type ResetBackend interface {
	Backend
	// TearDown removes existing workspaces before the clean-slate rebuild.
	TearDown() error
}

// Reset tears down existing workspaces, then rebuilds from config. Restore is
// intentionally skipped — this is a clean-slate rebuild, not a replay.
func Reset(b ResetBackend) error {
	if err := b.TearDown(); err != nil {
		return err
	}
	return Start(StartOptions{NoRestore: true}, b)
}
