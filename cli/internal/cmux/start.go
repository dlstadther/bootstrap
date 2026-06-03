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

// workspaceJSON is a single entry from cmux workspace list --json.
type workspaceJSON struct {
	Ref   string `json:"ref"`
	Title string `json:"title"`
}

// workspaceListJSON is the top-level response from cmux workspace list --json.
type workspaceListJSON struct {
	Workspaces []workspaceJSON `json:"workspaces"`
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
	if out, err := exec.Run("cmux", "ping"); err != nil {
		detail := out
		if detail == "" {
			detail = err.Error()
		}
		return fmt.Errorf("cmux is not running (%s): ensure cmux is installed and running", detail)
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
				exec.Run("cmux", "workspace", "close", wsRef) //nolint:errcheck
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
func findWorkspace(name string, exec Executor) string {
	out, err := exec.Run("cmux", "workspace", "list", "--json")
	if err != nil || out == "" {
		return ""
	}
	var result workspaceListJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return ""
	}
	for _, ws := range result.Workspaces {
		if ws.Title == name {
			return ws.Ref
		}
	}
	return ""
}

// listAllWorkspaceIDs returns the ref of every open workspace.
func listAllWorkspaceIDs(exec Executor) []string {
	out, err := exec.Run("cmux", "workspace", "list", "--json")
	if err != nil || out == "" {
		return nil
	}
	var result workspaceListJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil
	}
	ids := make([]string, 0, len(result.Workspaces))
	for _, ws := range result.Workspaces {
		ids = append(ids, ws.Ref)
	}
	return ids
}

// ResetOptions configures bs cmux reset behavior.
type ResetOptions struct {
	WorkspacesDir      string
	LocalWorkspacesDir string
	SkipWorkspaceID    string // if non-empty, skip closing this workspace (allows running from inside cmux)
}

// Reset closes all open cmux workspaces and rebuilds from JSON configs.
// restore-session is intentionally skipped — this is a clean-slate rebuild.
// If SkipWorkspaceID is set, that workspace is preserved (allows running from inside cmux).
func Reset(opts ResetOptions, exec Executor) error {
	if out, err := exec.Run("cmux", "ping"); err != nil {
		detail := out
		if detail == "" {
			detail = err.Error()
		}
		return fmt.Errorf("cmux is not running (%s): ensure cmux is installed and running", detail)
	}

	for _, id := range listAllWorkspaceIDs(exec) {
		if opts.SkipWorkspaceID != "" && id == opts.SkipWorkspaceID {
			continue
		}
		exec.Run("cmux", "workspace", "close", id) //nolint:errcheck
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

	wsOut, err := exec.Run("cmux", "workspace", "create", "--name", wc.Name, "--cwd", expandHome(wc.CWD))
	if err != nil {
		return fmt.Errorf("workspace create: %w", err)
	}
	// Output format is "OK <ref>"; strip the status prefix.
	wsID := strings.TrimPrefix(strings.TrimSpace(wsOut), "OK ")

	initialPaneID := firstPane(wsID, exec)

	for i, p := range wc.Panes {
		if i == 0 {
			if p.Command != "" {
				sendToPane(exec, wsID, p.Command, p.NoEnter)
			}
			continue
		}
		if _, err := exec.Run("cmux", "new-split", p.Split, "--workspace", wsID, "--focus", "true"); err != nil {
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
