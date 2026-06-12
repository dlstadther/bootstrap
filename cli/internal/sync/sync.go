package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// SyncMise runs mise install to sync tool versions.
func SyncMise(exec Executor) error {
	fmt.Println("syncing mise...")
	if _, err := exec.Run("mise", "install"); err != nil {
		return fmt.Errorf("mise install: %w", err)
	}
	return nil
}

// SyncBrew syncs Homebrew packages. When force is false, it runs brew bundle
// check first and skips the install if everything is already satisfied.
func SyncBrew(exec Executor, force bool) error {
	fmt.Println("syncing brew...")
	if !force {
		if _, err := exec.Run("brew", "bundle", "check", "--global"); err == nil {
			fmt.Println("brew already satisfied — skipping install")
			return nil
		}
	}
	if _, err := exec.Run("brew", "bundle", "install", "--global"); err != nil {
		return fmt.Errorf("brew bundle install: %w", err)
	}
	return nil
}

type claudeSettings struct {
	EnabledPlugins map[string]bool `json:"enabledPlugins"`
}

func readEnabledPlugins(settingsPath string) ([]string, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	var s claudeSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	var plugins []string
	for name, enabled := range s.EnabledPlugins {
		if enabled {
			plugins = append(plugins, name)
		}
	}
	sort.Strings(plugins)
	return plugins, nil
}

// SyncPlugins installs all enabled Claude plugins listed in settings.json.
// Each plugin is installed independently; one failure does not abort the rest.
func SyncPlugins(settingsPath string, exec Executor) error {
	fmt.Println("syncing plugins...")
	plugins, err := readEnabledPlugins(settingsPath)
	if err != nil {
		return err
	}
	var errs []error
	for _, plugin := range plugins {
		if _, err := exec.Run("claude", "plugin", "install", plugin); err != nil {
			fmt.Fprintf(os.Stderr, "  plugin %s: %v\n", plugin, err)
			errs = append(errs, fmt.Errorf("install %s: %w", plugin, err))
		}
	}
	return errors.Join(errs...)
}
