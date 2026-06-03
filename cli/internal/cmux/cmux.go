package cmux

import (
	"fmt"
	"path/filepath"
	"strings"
)

var AllowedAgents = []string{"claude", "codex", "gemini", "opencode", "pi"}

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// AddOptions configures a new cmux workspace.
type AddOptions struct {
	Name  string
	CWD   string
	Agent string
}

// Add creates a new cmux workspace with the 3-pane agent layout:
//   Left pane:    agent command staged (no Enter)
//   Top-right:    ls -al && bd ready (executed)
//   Bottom-right: lazygit (executed)
func Add(opts AddOptions, exec Executor) error {
	if opts.CWD == "" {
		return fmt.Errorf("--cwd is required")
	}
	if !isValidAgent(opts.Agent) {
		return fmt.Errorf("invalid agent %q; allowed: %s", opts.Agent, strings.Join(AllowedAgents, ", "))
	}

	if _, err := exec.Run("cmux", "ping"); err != nil {
		return fmt.Errorf("cmux is not running: ensure cmux is installed and running")
	}

	workspaceName := opts.Name
	if workspaceName == "" {
		workspaceName = filepath.Base(opts.CWD)
	}

	// Create workspace; the initial terminal opens at --cwd.
	wsOut, err := exec.Run("cmux", "workspace", "create", "--name", workspaceName, "--cwd", opts.CWD)
	if err != nil {
		return fmt.Errorf("workspace create: %w", err)
	}
	wsID := strings.TrimSpace(wsOut)

	// Capture the left pane ID before splitting so we can focus it later.
	leftPaneID := firstPane(wsID, exec)

	// Split right — new right pane is now the active surface.
	if _, err := exec.Run("cmux", withWS(wsID, "new-split", "right")...); err != nil {
		return fmt.Errorf("new-split right: %w", err)
	}

	// Top-right: cd + ls + bd ready (executed).
	send(exec, wsID, fmt.Sprintf("cd %s && ls -al && bd ready", shellQuote(opts.CWD)))
	sendKey(exec, wsID, "enter")

	// Split down — new bottom-right pane is now the active surface.
	if _, err := exec.Run("cmux", withWS(wsID, "new-split", "down")...); err != nil {
		return fmt.Errorf("new-split down: %w", err)
	}

	// Bottom-right: cd + lazygit (executed).
	send(exec, wsID, fmt.Sprintf("cd %s && lazygit", shellQuote(opts.CWD)))
	sendKey(exec, wsID, "enter")

	// Focus left pane, then stage agent command (no Enter).
	if leftPaneID != "" {
		exec.Run("cmux", withWS(wsID, "focus-pane", "--pane", leftPaneID)...) //nolint:errcheck
	}
	send(exec, wsID, buildAgentCmd(opts.Agent, opts.CWD))

	return nil
}

func isValidAgent(agent string) bool {
	for _, a := range AllowedAgents {
		if a == agent {
			return true
		}
	}
	return false
}

func buildAgentCmd(agent, cwd string) string {
	if agent == "claude" {
		return fmt.Sprintf("claude agents --cwd %s", shellQuote(cwd))
	}
	return agent
}

// withWS builds a cmux subcommand args slice, appending --workspace wsID if non-empty.
func withWS(wsID string, subcmd string, args ...string) []string {
	result := append([]string{subcmd}, args...)
	if wsID != "" {
		result = append(result, "--workspace", wsID)
	}
	return result
}

func send(exec Executor, wsID, text string) {
	args := []string{"send"}
	if wsID != "" {
		args = append(args, "--workspace", wsID)
	}
	args = append(args, text)
	exec.Run("cmux", args...) //nolint:errcheck
}

func sendKey(exec Executor, wsID, key string) {
	args := []string{"send-key"}
	if wsID != "" {
		args = append(args, "--workspace", wsID)
	}
	args = append(args, key)
	exec.Run("cmux", args...) //nolint:errcheck
}

// firstPane returns the ref of the first pane in the workspace by parsing list-panes output.
func firstPane(wsID string, exec Executor) string {
	args := []string{"list-panes"}
	if wsID != "" {
		args = append(args, "--workspace", wsID)
	}
	out, err := exec.Run("cmux", args...)
	if err != nil || out == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0])
}

// shellQuote wraps a path in single quotes, escaping any existing single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
