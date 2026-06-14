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
	Command     string `json:"command"`
	Split       string `json:"split"`        // "right" | "down" | "left" | "up"; empty for first pane
	NoEnter     bool   `json:"no_enter"`     // stage without executing
	SizePercent int    `json:"size_percent"` // percentage this pane occupies (0 = 50% default)
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

// layoutSurface is a terminal surface entry in a cmux layout.
type layoutSurface struct {
	Type string `json:"type"`
}

// layoutPane wraps a list of surfaces in a pane node.
type layoutPane struct {
	Surfaces []layoutSurface `json:"surfaces"`
}

// layoutNode is a node in the cmux layout tree: either a leaf pane or a split.
type layoutNode struct {
	Pane      *layoutPane  `json:"pane,omitempty"`
	Direction string       `json:"direction,omitempty"`
	Split     float64      `json:"split,omitempty"`
	Children  []layoutNode `json:"children,omitempty"`
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

	cwd, err := expandHome(wc.CWD)
	if err != nil {
		return err
	}
	args := []string{"workspace", "create", "--name", wc.Name, "--cwd", cwd}
	if len(wc.Panes) > 1 {
		layout := buildLayout(wc.Panes, 0)
		data, err := json.Marshal(layout)
		if err != nil {
			return fmt.Errorf("build layout: %w", err)
		}
		args = append(args, "--layout", string(data))
	}

	wsOut, err := exec.Run("cmux", args...)
	if err != nil {
		return fmt.Errorf("workspace create: %w", err)
	}
	wsID, err := parseWorkspaceRef(wsOut)
	if err != nil {
		return err
	}

	// Get pane IDs, then resolve each pane's first surface for targeted sends.
	// send --surface requires surface refs; pane refs are not valid surface IDs.
	paneIDs := listPaneIDs(wsID, exec)
	surfaceIDs := listSurfaceIDsForPanes(wsID, paneIDs, exec)

	for i, p := range wc.Panes {
		if p.Command == "" {
			continue
		}
		sendArgs := []string{"send", "--workspace", wsID}
		if i < len(surfaceIDs) && surfaceIDs[i] != "" {
			sendArgs = append(sendArgs, "--surface", surfaceIDs[i])
		}
		sendArgs = append(sendArgs, p.Command)
		exec.Run("cmux", sendArgs...) //nolint:errcheck

		if !p.NoEnter {
			keyArgs := []string{"send-key", "--workspace", wsID}
			if i < len(surfaceIDs) && surfaceIDs[i] != "" {
				keyArgs = append(keyArgs, "--surface", surfaceIDs[i])
			}
			keyArgs = append(keyArgs, "enter")
			exec.Run("cmux", keyArgs...) //nolint:errcheck
		}
	}

	// Return focus to the first (initial) pane.
	if len(paneIDs) > 0 {
		exec.Run("cmux", "focus-pane", "--pane", paneIDs[0], "--workspace", wsID) //nolint:errcheck
	}

	return nil
}

// buildLayout recursively constructs a cmux layout tree from pane specs.
// Each split pane (index > 0) extends from the previous pane's position,
// mirroring the sequential split model. Only "right" and "down" are common;
// "left" and "up" place the new pane as the first child.
//
// Split (cmux's fraction) always describes the FIRST child's share, but the new
// pane lands in a different child slot depending on direction, so SizePercent
// must be flipped to match:
//   - "right"/"down": new pane is the SECOND child, so the first child keeps the
//     remainder → split = 1 - sp.
//   - "left"/"up":    new pane is the FIRST child, so it gets sp directly →
//     split = sp.
// Without this flip, --size-percent would size the wrong pane for left/up.
func buildLayout(panes []PaneSpec, idx int) layoutNode {
	leaf := layoutNode{Pane: &layoutPane{Surfaces: []layoutSurface{{Type: "terminal"}}}}
	if idx >= len(panes)-1 {
		return leaf
	}

	next := panes[idx+1]
	var direction string
	switch next.Split {
	case "right", "left":
		direction = "horizontal"
	default: // "down", "up"
		direction = "vertical"
	}

	// split is the fraction given to the FIRST child.
	split := 0.5
	if sp := next.SizePercent; sp > 0 && sp < 100 {
		if next.Split == "right" || next.Split == "down" {
			// New pane is the second child; first child gets the remainder.
			split = 1.0 - float64(sp)/100.0
		} else {
			// New pane is the first child; it gets sp%.
			split = float64(sp) / 100.0
		}
	}

	rest := buildLayout(panes, idx+1)

	if next.Split == "left" || next.Split == "up" {
		return layoutNode{Direction: direction, Split: split, Children: []layoutNode{rest, leaf}}
	}
	return layoutNode{Direction: direction, Split: split, Children: []layoutNode{leaf, rest}}
}

// listPaneIDs returns the ordered refs of all panes in a workspace.
// cmux list-panes outputs lines like "* pane:1  [1 surface]  [focused]" or
// "  pane:2  [1 surface]"; extract the token starting with "pane:".
func listPaneIDs(wsID string, exec Executor) []string {
	out, err := exec.Run("cmux", "list-panes", "--workspace", wsID)
	if err != nil || out == "" {
		return nil
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		for _, f := range strings.Fields(line) {
			if strings.HasPrefix(f, "pane:") {
				ids = append(ids, f)
				break
			}
		}
	}
	return ids
}

// listSurfaceIDsForPanes returns the first surface ref for each pane, in pane order.
// cmux send --surface requires surface refs (surface:N), not pane refs (pane:N).
// cmux list-pane-surfaces outputs lines like "* surface:1  Terminal  [selected]";
// extract the token starting with "surface:".
func listSurfaceIDsForPanes(wsID string, paneIDs []string, exec Executor) []string {
	surfaceIDs := make([]string, len(paneIDs))
	for i, paneID := range paneIDs {
		out, err := exec.Run("cmux", "list-pane-surfaces", "--workspace", wsID, "--pane", paneID)
		if err != nil || out == "" {
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			for _, f := range strings.Fields(line) {
				if strings.HasPrefix(f, "surface:") {
					surfaceIDs[i] = f
					break
				}
			}
			if surfaceIDs[i] != "" {
				break
			}
		}
	}
	return surfaceIDs
}

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand %q: home lookup failed: %w", path, err)
		}
		return home + path[1:], nil
	}
	return path, nil
}

// parseWorkspaceRef extracts the workspace ref from cmux's "OK <ref>" output.
// Validates the ref is present; a bare "OK" (no ref) is an error rather than
// silently yielding "OK" as the id and breaking downstream cmux calls.
func parseWorkspaceRef(out string) (string, error) {
	fields := strings.SplitN(strings.TrimSpace(out), " ", 2)
	if len(fields) != 2 || fields[0] != "OK" || strings.TrimSpace(fields[1]) == "" {
		return "", fmt.Errorf("unexpected workspace create output: %q", out)
	}
	return strings.TrimSpace(fields[1]), nil
}
