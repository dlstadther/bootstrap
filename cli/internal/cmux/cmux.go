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

// agentLayout is the 3-pane layout for bs cmux add:
//   Left  60%: agent command staged (no Enter)
//   Right 40%, split 50/50:
//     Top-right:    ls -al && bd ready (executed)
//     Bottom-right: lazygit (executed)
const agentLayout = `{"direction":"horizontal","split":0.6,"children":[` +
	`{"pane":{"surfaces":[{"type":"terminal"}]}},` +
	`{"direction":"vertical","split":0.5,"children":[` +
	`{"pane":{"surfaces":[{"type":"terminal"}]}},` +
	`{"pane":{"surfaces":[{"type":"terminal"}]}}` +
	`]}]}`

// Add creates a new cmux workspace with the 3-pane agent layout.
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

	wsOut, err := exec.Run("cmux", "workspace", "create", "--name", workspaceName, "--cwd", opts.CWD, "--layout", agentLayout)
	if err != nil {
		return fmt.Errorf("workspace create: %w", err)
	}
	wsID := strings.TrimPrefix(strings.TrimSpace(wsOut), "OK ")

	paneIDs := listPaneIDs(wsID, exec)
	pane := func(i int) string {
		if i < len(paneIDs) {
			return paneIDs[i]
		}
		return ""
	}

	// Top-right (pane 1): cd + ls + bd ready.
	if id := pane(1); id != "" {
		exec.Run("cmux", "focus-pane", "--pane", id, "--workspace", wsID) //nolint:errcheck
		exec.Run("cmux", "send", "--workspace", wsID,                     //nolint:errcheck
			fmt.Sprintf("cd %s && ls -al && bd ready", shellQuote(opts.CWD)))
		exec.Run("cmux", "send-key", "--workspace", wsID, "enter") //nolint:errcheck
	}

	// Bottom-right (pane 2): cd + lazygit.
	if id := pane(2); id != "" {
		exec.Run("cmux", "focus-pane", "--pane", id, "--workspace", wsID) //nolint:errcheck
		exec.Run("cmux", "send", "--workspace", wsID,                     //nolint:errcheck
			fmt.Sprintf("cd %s && lazygit", shellQuote(opts.CWD)))
		exec.Run("cmux", "send-key", "--workspace", wsID, "enter") //nolint:errcheck
	}

	// Left pane (pane 0): focus and stage agent command (no Enter).
	if id := pane(0); id != "" {
		exec.Run("cmux", "focus-pane", "--pane", id, "--workspace", wsID) //nolint:errcheck
		exec.Run("cmux", "send", "--workspace", wsID, buildAgentCmd(opts.Agent, opts.CWD)) //nolint:errcheck
	}

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

// shellQuote wraps a path in single quotes, escaping any existing single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
