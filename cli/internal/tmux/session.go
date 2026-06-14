package tmux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"gopkg.in/yaml.v3"
)

// SessionConfig represents a tmux session defined in a YAML file.
// Format is compatible with TMUXinator's config schema.
type SessionConfig struct {
	Name    string         `yaml:"name"`
	Root    string         `yaml:"root"`
	Windows []WindowConfig `yaml:"-"`
}

// WindowConfig represents a single tmux window within a session.
type WindowConfig struct {
	Name            string
	Root            string
	Layout          string
	MainPanePercent int // if > 0, set main-pane-width to this % after applying layout (tmux >= 3.1)
	Panes           []PaneConfig
}

// PaneConfig represents a single pane within a window.
type PaneConfig struct {
	Command string
	NoEnter bool // if true, send keys without pressing Enter (stage only)
}

// StartOptions configures bs tmux start behavior.
type StartOptions struct {
	NoRestore        bool
	Override         bool
	SessionsDir      string
	LocalSessionsDir string // optional; loaded after SessionsDir (like bootstrap.local.sh)
	ResurrectPath    string
	AfterRestoreWait time.Duration
}

// ParseSession parses a TMUXinator-compatible YAML session config.
func ParseSession(data []byte) (*SessionConfig, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return parseSessionNode(&doc)
}

// LoadSessions reads all *.yaml files from dir and parses them as SessionConfigs.
func LoadSessions(dir string) ([]SessionConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []SessionConfig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		s, err := ParseSession(data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		sessions = append(sessions, *s)
	}
	return sessions, nil
}

// Start implements the bs tmux start flow:
// 1. Ensure tmux is running
// 2. Optionally restore via tmux-resurrect
// 3. Optionally kill conflicting sessions (--override)
// 4. Create sessions from YAML configs
func Start(opts StartOptions, exec iexec.Executor) error {
	if _, err := exec.Run("tmux", "info"); err != nil {
		if _, err2 := exec.Run("tmux", "new-session", "-d", "-s", "main"); err2 != nil {
			return fmt.Errorf("start tmux: %w", err2)
		}
	}

	if !opts.NoRestore {
		if _, err := os.Stat(opts.ResurrectPath); err != nil {
			return fmt.Errorf("tmux-resurrect not found at %s", opts.ResurrectPath)
		}
		if _, err := exec.Run("tmux", "run-shell", opts.ResurrectPath); err != nil {
			return fmt.Errorf("resurrect restore: %w", err)
		}
		if opts.AfterRestoreWait > 0 {
			time.Sleep(opts.AfterRestoreWait)
		}
	}

	sessions, err := LoadSessions(opts.SessionsDir)
	if err != nil {
		return fmt.Errorf("load sessions: %w", err)
	}
	if opts.LocalSessionsDir != "" {
		local, err := LoadSessions(opts.LocalSessionsDir)
		if err != nil {
			return fmt.Errorf("load local sessions: %w", err)
		}
		sessions = append(sessions, local...)
	}

	if opts.Override {
		for _, s := range sessions {
			if sessionExists(s.Name, exec) {
				exec.Run("tmux", "kill-session", "-t", s.Name) //nolint:errcheck
			}
		}
	}

	for _, s := range sessions {
		if err := createSessionFromConfig(s, exec); err != nil {
			return fmt.Errorf("session %s: %w", s.Name, err)
		}
	}
	return nil
}

func createSessionFromConfig(s SessionConfig, exec iexec.Executor) error {
	// If the session already exists, leave it untouched.
	// tmux-resurrect may have restored its layout — adding panes would create duplicates.
	if sessionExists(s.Name, exec) {
		return nil
	}

	for i, w := range s.Windows {
		root := coalesce(w.Root, s.Root)
		if i == 0 {
			if !sessionExists(s.Name, exec) {
				if _, err := exec.Run("tmux", "new-session", "-d", "-s", s.Name, "-n", w.Name, "-c", expandHome(root)); err != nil {
					return fmt.Errorf("new-session: %w", err)
				}
			}
		} else {
			if !windowExists(s.Name, w.Name, exec) {
				if _, err := exec.Run("tmux", "new-window", "-t", s.Name, "-n", w.Name, "-c", expandHome(root)); err != nil {
					return fmt.Errorf("new-window %s: %w", w.Name, err)
				}
			}
		}

		if w.Layout != "" {
			exec.Run("tmux", "select-layout", "-t", s.Name+":"+w.Name, w.Layout) //nolint:errcheck
		}

		if err := createPanesFromConfig(s.Name, w, root, exec); err != nil {
			return err
		}
	}
	return nil
}

func createPanesFromConfig(session string, w WindowConfig, defaultRoot string, exec iexec.Executor) error {
	target := session + ":" + w.Name
	root := coalesce(w.Root, defaultRoot)

	for i, p := range w.Panes {
		if i > 0 {
			if _, err := exec.Run("tmux", "split-window", "-t", target, "-c", expandHome(root)); err != nil {
				return fmt.Errorf("split-window: %w", err)
			}
			if w.Layout != "" {
				exec.Run("tmux", "select-layout", "-t", target, w.Layout) //nolint:errcheck
			}
		}
		// Target the active pane in the window — after split-window the new pane
		// is always active, so this correctly follows pane creation order without
		// hardcoding numeric indices (which break with pane-base-index != 0).
		if p.Command != "" {
			if p.NoEnter {
				exec.Run("tmux", "send-keys", "-t", target, p.Command) //nolint:errcheck
			} else {
				exec.Run("tmux", "send-keys", "-t", target, p.Command, "Enter") //nolint:errcheck
			}
		}
	}

	// Apply main pane size constraint after all panes are created.
	// Uses percentage syntax — requires tmux >= 3.1.
	if w.Layout != "" && w.MainPanePercent > 0 {
		exec.Run("tmux", "set-window-option", "-t", target, "main-pane-width", fmt.Sprintf("%d%%", w.MainPanePercent)) //nolint:errcheck
		exec.Run("tmux", "select-layout", "-t", target, w.Layout)                                                      //nolint:errcheck
	}

	return nil
}

func parseWindowNode(node *yaml.Node) (WindowConfig, error) {
	// Each window item is a mapping node: {window_name: value}
	if node.Kind != yaml.MappingNode || len(node.Content) < 2 {
		return WindowConfig{}, fmt.Errorf("invalid window node (kind=%d, len=%d)", node.Kind, len(node.Content))
	}
	name := node.Content[0].Value
	valueNode := node.Content[1]
	wc := WindowConfig{Name: name}

	switch valueNode.Kind {
	case yaml.ScalarNode:
		wc.Panes = []PaneConfig{{Command: valueNode.Value}}
	case yaml.MappingNode:
		var detail struct {
			Root            string      `yaml:"root"`
			Layout          string      `yaml:"layout"`
			MainPanePercent int         `yaml:"main_pane_percent"`
			Panes           []yaml.Node `yaml:"panes"`
		}
		if err := valueNode.Decode(&detail); err != nil {
			return WindowConfig{}, err
		}
		wc.Root = detail.Root
		wc.Layout = detail.Layout
		wc.MainPanePercent = detail.MainPanePercent
		for i := range detail.Panes {
			pc, err := parsePaneNode(&detail.Panes[i])
			if err != nil {
				return WindowConfig{}, err
			}
			wc.Panes = append(wc.Panes, pc)
		}
	default:
		return WindowConfig{}, fmt.Errorf("unexpected window value kind: %d", valueNode.Kind)
	}
	return wc, nil
}

func parsePaneNode(node *yaml.Node) (PaneConfig, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return PaneConfig{Command: node.Value}, nil
	case yaml.MappingNode:
		var p struct {
			Command string `yaml:"command"`
			NoEnter bool   `yaml:"no_enter"`
		}
		if err := node.Decode(&p); err != nil {
			return PaneConfig{}, err
		}
		return PaneConfig{Command: p.Command, NoEnter: p.NoEnter}, nil
	default:
		return PaneConfig{}, fmt.Errorf("invalid pane node kind: %d", node.Kind)
	}
}

// ParseSession parses YAML by walking the document node directly,
// so window list items (map entries) are correctly identified as MappingNodes.
func parseSessionNode(doc *yaml.Node) (*SessionConfig, error) {
	// doc.Kind == DocumentNode, doc.Content[0] is the root mapping
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, fmt.Errorf("unexpected YAML structure")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping at root")
	}

	s := &SessionConfig{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i].Value
		val := root.Content[i+1]
		switch key {
		case "name":
			s.Name = val.Value
		case "root":
			s.Root = val.Value
		case "windows":
			if val.Kind != yaml.SequenceNode {
				return nil, fmt.Errorf("windows must be a sequence")
			}
			for _, item := range val.Content {
				wc, err := parseWindowNode(item)
				if err != nil {
					return nil, err
				}
				s.Windows = append(s.Windows, wc)
			}
		}
	}
	return s, nil
}

// ResetOptions configures bs tmux reset behavior.
type ResetOptions struct {
	SessionsDir      string
	LocalSessionsDir string
}

// Reset kills the tmux server and rebuilds sessions from YAML configs.
// Resurrect restore is intentionally skipped — this is a clean-slate rebuild.
func Reset(opts ResetOptions, exec iexec.Executor) error {
	exec.Run("tmux", "kill-server") //nolint:errcheck — ignore if server wasn't running
	return Start(StartOptions{
		NoRestore:        true,
		SessionsDir:      opts.SessionsDir,
		LocalSessionsDir: opts.LocalSessionsDir,
	}, exec)
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
