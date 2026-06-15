package tmux

import (
	"fmt"
	"path/filepath"
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	"github.com/dlstadther/bootstrap/cli/internal/workspace"
)

// AllowedAgents re-exports the canonical agent list so callers (and the cmd
// flag help) can reference it as tmux.AllowedAgents.
var AllowedAgents = workspace.AllowedAgents

// AddOptions configures a new tmux workspace window.
type AddOptions struct {
	Name  string
	CWD   string
	Agent string
}

// Add creates (or joins) a tmux session and opens a workspace window.
func Add(opts AddOptions, exec iexec.Executor) error {
	if opts.CWD == "" {
		return fmt.Errorf("--cwd is required")
	}
	if err := workspace.ValidateAgent(opts.Agent); err != nil {
		return err
	}

	dirname := filepath.Base(opts.CWD)
	sessionName := dirname
	windowName := opts.Name
	if windowName == "" {
		windowName = dirname
	}

	// Check if tmux is running.
	if _, err := exec.Run("tmux", "info"); err != nil {
		return fmt.Errorf("tmux is not running")
	}

	hasSession := sessionExists(sessionName, exec)

	if !hasSession {
		if _, err := exec.Run("tmux", "new-session", "-d", "-s", sessionName, "-n", windowName, "-c", opts.CWD); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
	} else {
		windowName = ensureUniqueWindowName(sessionName, windowName, opts.Name == "", exec)
		if _, err := exec.Run("tmux", "new-window", "-t", sessionName, "-n", windowName, "-c", opts.CWD); err != nil {
			return fmt.Errorf("new window: %w", err)
		}
	}

	if err := createPanes(sessionName, windowName, opts.CWD, opts.Agent, exec); err != nil {
		return err
	}

	if !hasSession {
		_, _ = exec.Run("tmux", "switch-client", "-t", sessionName)
	} else {
		_, _ = exec.Run("tmux", "select-window", "-t", sessionName+":"+windowName)
	}

	return nil
}

func sessionExists(name string, exec iexec.Executor) bool {
	_, err := exec.Run("tmux", "has-session", "-t", name)
	return err == nil
}

// ensureUniqueWindowName returns base unchanged when the user supplied an
// explicit name (autoName=false). For auto-named windows it appends -2, -3, …
// until the name is free, avoiding collisions with existing windows.
func ensureUniqueWindowName(session, base string, autoName bool, exec iexec.Executor) string {
	if !autoName {
		return base
	}
	candidate := base
	n := 2
	for windowExists(session, candidate, exec) {
		candidate = fmt.Sprintf("%s-%d", base, n)
		n++
	}
	return candidate
}

func windowExists(session, name string, exec iexec.Executor) bool {
	out, err := exec.Run("tmux", "list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return false
	}
	for _, w := range strings.Split(strings.TrimSpace(out), "\n") {
		if w == name {
			return true
		}
	}
	return false
}

func createPanes(session, window, cwd, agent string, exec iexec.Executor) error {
	target := session + ":" + window

	// Capture the left pane's ID before any splits. Using the pane ID (e.g. %42)
	// is immune to pane-base-index, unlike hardcoded .0 which breaks when
	// pane-base-index = 1 (a common setting).
	leftPaneID, _ := exec.Run("tmux", "display-message", "-t", target, "-p", "#{pane_id}")

	// Split right column (~40%) — new right pane is now active
	if _, err := exec.Run("tmux", "split-window", "-h", "-p", "40", "-t", target, "-c", cwd); err != nil {
		return fmt.Errorf("split-window horizontal: %w", err)
	}

	// Top-right: ls + bd ready — send directly to the active right pane
	_, _ = exec.Run("tmux", "send-keys", "-t", target, "ls -al && bd ready", "Enter")

	// Bottom-right: lazygit — one vertical split of the right pane
	if _, err := exec.Run("tmux", "split-window", "-v", "-t", target, "-c", cwd); err != nil {
		return fmt.Errorf("split-window vertical (bottom-right): %w", err)
	}
	_, _ = exec.Run("tmux", "send-keys", "-t", target, "lazygit", "Enter")

	// Left pane: stage agent (no Enter — user stays in control).
	// Fall back to target (active pane) if display-message didn't return an ID.
	leftTarget := leftPaneID
	if leftTarget == "" {
		leftTarget = target
	}
	agentCmd := agent
	if agent == "claude" {
		agentCmd = fmt.Sprintf("claude agents --cwd %s", cwd)
	}
	_, _ = exec.Run("tmux", "select-pane", "-t", leftTarget)
	_, _ = exec.Run("tmux", "send-keys", "-t", leftTarget, agentCmd, "")

	return nil
}
