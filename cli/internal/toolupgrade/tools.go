package toolupgrade

import (
	"encoding/json"
	"fmt"
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
)

// Registry returns the ordered set of top-level tools to manage.
func Registry() []Tool {
	return []Tool{brewTool{}, claudeTool{}, opencodeTool{}}
}

// --- helpers ---

// firstLine returns the first line of s (without the trailing newline).
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// firstField returns the first whitespace-separated token of s.
func firstField(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// githubLatestTag fetches the latest release tag for owner/repo via the GitHub
// API and strips a leading "v".
func githubLatestTag(exec iexec.LookPathExecutor, repo string) (string, error) {
	out, err := exec.Run("curl", "-fsSL", "https://api.github.com/repos/"+repo+"/releases/latest")
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal([]byte(out), &rel); err != nil {
		return "", fmt.Errorf("parse %s release: %w", repo, err)
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

// --- brew ---

type brewTool struct{}

func (brewTool) Name() string { return "brew" }
func (brewTool) Installed(exec iexec.LookPathExecutor) bool {
	_, err := exec.LookPath("brew")
	return err == nil
}
func (brewTool) CurrentVersion(exec iexec.LookPathExecutor) (string, error) {
	out, err := exec.Run("brew", "--version")
	if err != nil {
		return "", err
	}
	// "Homebrew 5.1.14" → "5.1.14"
	fields := strings.Fields(firstLine(out))
	if len(fields) < 2 {
		return "", fmt.Errorf("unexpected brew --version output: %q", out)
	}
	return fields[1], nil
}
func (brewTool) LatestVersion(exec iexec.LookPathExecutor) (string, error) {
	return githubLatestTag(exec, "Homebrew/brew")
}
func (brewTool) Upgrade(exec iexec.LookPathExecutor) error {
	if _, err := exec.Run("brew", "update"); err != nil {
		return fmt.Errorf("brew update: %w", err)
	}
	return nil
}

// --- claude ---

type claudeTool struct{}

func (claudeTool) Name() string { return "claude" }
func (claudeTool) Installed(exec iexec.LookPathExecutor) bool {
	_, err := exec.LookPath("claude")
	return err == nil
}
func (claudeTool) CurrentVersion(exec iexec.LookPathExecutor) (string, error) {
	out, err := exec.Run("claude", "--version")
	if err != nil {
		return "", err
	}
	// "2.1.165 (Claude Code)" → "2.1.165"
	return firstField(out), nil
}
func (claudeTool) LatestVersion(iexec.LookPathExecutor) (string, error) {
	// No clean pre-check endpoint; `claude update` determines latest on apply.
	return "", nil
}
func (claudeTool) Upgrade(exec iexec.LookPathExecutor) error {
	if _, err := exec.Run("claude", "update"); err != nil {
		return fmt.Errorf("claude update: %w", err)
	}
	return nil
}

// --- opencode ---

type opencodeTool struct{}

func (opencodeTool) Name() string { return "opencode" }
func (opencodeTool) Installed(exec iexec.LookPathExecutor) bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}
func (opencodeTool) CurrentVersion(exec iexec.LookPathExecutor) (string, error) {
	out, err := exec.Run("opencode", "--version")
	if err != nil {
		return "", err
	}
	// "1.2.10" → "1.2.10"
	return firstField(out), nil
}
func (opencodeTool) LatestVersion(exec iexec.LookPathExecutor) (string, error) {
	return githubLatestTag(exec, "sst/opencode")
}
func (opencodeTool) Upgrade(exec iexec.LookPathExecutor) error {
	if _, err := exec.Run("opencode", "upgrade"); err != nil {
		return fmt.Errorf("opencode upgrade: %w", err)
	}
	return nil
}
