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
	surfaceIDs := listSurfaceIDsForPanes(wsID, paneIDs, exec)
	surface := func(i int) string {
		if i < len(surfaceIDs) {
			return surfaceIDs[i]
		}
		return ""
	}

	// Top-right (pane 1): cd + ls + bd ready.
	if id := surface(1); id != "" {
		exec.Run("cmux", "send", "--workspace", wsID, "--surface", id, //nolint:errcheck
			fmt.Sprintf("cd %s && ls -al && bd ready", shellQuote(opts.CWD)))
		exec.Run("cmux", "send-key", "--workspace", wsID, "--surface", id, "enter") //nolint:errcheck
	}

	// Bottom-right (pane 2): cd + lazygit.
	if id := surface(2); id != "" {
		exec.Run("cmux", "send", "--workspace", wsID, "--surface", id, //nolint:errcheck
			fmt.Sprintf("cd %s && lazygit", shellQuote(opts.CWD)))
		exec.Run("cmux", "send-key", "--workspace", wsID, "--surface", id, "enter") //nolint:errcheck
	}

	// Left pane (pane 0): stage agent command (no Enter), then focus it.
	agentArgs := []string{"send", "--workspace", wsID}
	if id := surface(0); id != "" {
		agentArgs = append(agentArgs, "--surface", id)
	}
	agentArgs = append(agentArgs, buildAgentCmd(opts.Agent, opts.CWD))
	exec.Run("cmux", agentArgs...) //nolint:errcheck

	// Focus the left pane so the user lands there to trigger the agent.
	if len(paneIDs) > 0 {
		exec.Run("cmux", "focus-pane", "--pane", paneIDs[0], "--workspace", wsID) //nolint:errcheck
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
