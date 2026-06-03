package cmux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceConfig defines a cmux workspace to create.
type WorkspaceConfig struct {
	Name  string     `json:"name"`
	CWD   string     `json:"cwd"`
	Panes []PaneSpec `json:"panes"`
}

// PaneSpec defines a single pane within a workspace.
// The first pane (index 0) is the workspace's initial pane.
// Subsequent panes must specify a Split direction.
type PaneSpec struct {
	Command string `json:"command"`
	Split   string `json:"split"`    // "right" | "down" | "left" | "up"; empty for first pane
	NoEnter bool   `json:"no_enter"` // stage without executing
}

// StartOptions configures bs cmux start behavior.
type StartOptions struct {
	NoRestore          bool
	Override           bool
	WorkspacesDir      string
	LocalWorkspacesDir string // optional; loaded after WorkspacesDir
}

// LoadWorkspaces reads all *.json files from dir and parses them as WorkspaceConfigs.
func LoadWorkspaces(dir string) ([]WorkspaceConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var workspaces []WorkspaceConfig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var wc WorkspaceConfig
		if err := json.Unmarshal(data, &wc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		workspaces = append(workspaces, wc)
	}
	return workspaces, nil
}

// Start implements the bs cmux start flow:
//  1. Ensure cmux is running
//  2. Optionally restore via cmux restore-session
//  3. Optionally close conflicting workspaces (--override)
//  4. Create workspaces from JSON configs
func Start(opts StartOptions, exec Executor) error {
	if _, err := exec.Run("cmux", "ping"); err != nil {
		return fmt.Errorf("cmux is not running: ensure cmux is installed and running")
	}

	if !opts.NoRestore {
		if _, err := exec.Run("cmux", "restore-session"); err != nil {
			return fmt.Errorf("restore-session: %w", err)
		}
	}

	workspaces, err := LoadWorkspaces(opts.WorkspacesDir)
	if err != nil {
		return fmt.Errorf("load workspaces: %w", err)
	}
	if opts.LocalWorkspacesDir != "" {
		local, err := LoadWorkspaces(opts.LocalWorkspacesDir)
		if err != nil {
			return fmt.Errorf("load local workspaces: %w", err)
		}
		workspaces = append(workspaces, local...)
	}

	if opts.Override {
		for _, w := range workspaces {
			if wsRef := findWorkspace(w.Name, exec); wsRef != "" {
				exec.Run("cmux", "close-workspace", "--workspace", wsRef) //nolint:errcheck
			}
		}
	}

	for _, w := range workspaces {
		if err := createWorkspaceFromConfig(w, exec); err != nil {
			return fmt.Errorf("workspace %s: %w", w.Name, err)
		}
	}
	return nil
}

// findWorkspace returns the workspace ref if a workspace named name exists, or "".
// cmux list-workspaces output is expected to have one workspace per line; the
// name appears as a whitespace-separated token. Adjust if the actual format differs.
func findWorkspace(name string, exec Executor) string {
	out, err := exec.Run("cmux", "list-workspaces")
	if err != nil || out == "" {
		return ""
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		for _, field := range strings.Fields(line) {
			if field == name {
				// Return the first field (workspace ref) from the line.
				return strings.Fields(line)[0]
			}
		}
	}
	return ""
}

// listAllWorkspaceIDs returns the ref of every open workspace (first token per line).
func listAllWorkspaceIDs(exec Executor) []string {
	out, err := exec.Run("cmux", "list-workspaces")
	if err != nil || out == "" {
		return nil
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			ids = append(ids, fields[0])
		}
	}
	return ids
}

// ResetOptions configures bs cmux reset behavior.
type ResetOptions struct {
	WorkspacesDir      string
	LocalWorkspacesDir string
}

// Reset closes all open cmux workspaces and rebuilds from JSON configs.
// restore-session is intentionally skipped — this is a clean-slate rebuild.
func Reset(opts ResetOptions, exec Executor) error {
	if _, err := exec.Run("cmux", "ping"); err != nil {
		return fmt.Errorf("cmux is not running: ensure cmux is installed and running")
	}

	for _, id := range listAllWorkspaceIDs(exec) {
		exec.Run("cmux", "close-workspace", "--workspace", id) //nolint:errcheck
	}

	return Start(StartOptions{
		NoRestore:          true,
		WorkspacesDir:      opts.WorkspacesDir,
		LocalWorkspacesDir: opts.LocalWorkspacesDir,
	}, exec)
}

func createWorkspaceFromConfig(wc WorkspaceConfig, exec Executor) error {
	// Leave existing workspaces untouched; restore-session may have recreated them.
	if findWorkspace(wc.Name, exec) != "" {
		return nil
	}

	wsOut, err := exec.Run("cmux", "new-workspace", "--name", wc.Name, "--cwd", expandHome(wc.CWD))
	if err != nil {
		return fmt.Errorf("new-workspace: %w", err)
	}
	wsID := strings.TrimSpace(wsOut)

	initialPaneID := firstPane(wsID, exec)

	for i, p := range wc.Panes {
		if i == 0 {
			if p.Command != "" {
				sendToPane(exec, wsID, p.Command, p.NoEnter)
			}
			continue
		}
		if _, err := exec.Run("cmux", "new-split", p.Split, "--workspace", wsID); err != nil {
			return fmt.Errorf("new-split %s: %w", p.Split, err)
		}
		if p.Command != "" {
			sendToPane(exec, wsID, p.Command, p.NoEnter)
		}
	}

	if initialPaneID != "" {
		exec.Run("cmux", "focus-pane", "--pane", initialPaneID, "--workspace", wsID) //nolint:errcheck
	}

	return nil
}

func sendToPane(exec Executor, wsID, text string, noEnter bool) {
	send(exec, wsID, text)
	if !noEnter {
		sendKey(exec, wsID, "enter")
	}
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
