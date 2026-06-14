package pluginupgrade

import (
	"fmt"
	"strings"
)

// Plugin is an installed, enabled Claude Code plugin.
type Plugin struct {
	name    string
	version string // empty when version reported as "unknown"
}

func (p Plugin) Name() string                              { return p.name }
func (p Plugin) Installed(_ Executor) bool                 { return true }
func (p Plugin) CurrentVersion(_ Executor) (string, error) { return p.version, nil }
func (p Plugin) Upgrade(exec Executor) error {
	if _, err := exec.Run("claude", "plugins", "update", p.name); err != nil {
		return fmt.Errorf("claude plugins update %s: %w", p.name, err)
	}
	return nil
}

// Discover runs `claude plugins list` and returns only enabled plugins.
func Discover(exec Executor) ([]Tool, error) {
	out, err := exec.Run("claude", "plugins", "list")
	if err != nil {
		return nil, fmt.Errorf("claude plugins list: %w", err)
	}
	return ParsePluginList(out), nil
}

// ParsePluginList parses the output of `claude plugins list` and returns only
// enabled plugins. Exported for testing.
func ParsePluginList(output string) []Tool {
	var tools []Tool
	var cur *Plugin

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "❯ ") {
			cur = &Plugin{name: strings.TrimPrefix(line, "❯ ")}
			continue
		}
		if cur == nil {
			continue
		}
		if strings.HasPrefix(line, "Version: ") {
			v := strings.TrimPrefix(line, "Version: ")
			if v != "unknown" {
				cur.version = v
			}
			continue
		}
		if line == "Status: ✔ enabled" {
			tools = append(tools, *cur)
		}
	}
	return tools
}
