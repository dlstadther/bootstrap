package brew

import (
	"errors"
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// BrewfilePath returns the active Brewfile path for repoPath.
// Prefers hosts/<machine>/.Brewfile; falls back to dotfiles/.Brewfile.
func BrewfilePath(repoPath string) (string, error) {
	machine, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("hostname: %w", err)
	}
	if idx := strings.Index(machine, "."); idx >= 0 {
		machine = machine[:idx]
	}
	host := filepath.Join(repoPath, "hosts", machine, ".Brewfile")
	if _, err := os.Stat(host); err == nil {
		return host, nil
	}
	return filepath.Join(repoPath, "dotfiles", ".Brewfile"), nil
}

// Sync shows drift between the live machine state and the repo Brewfile.
// repoBrewfile is the absolute path to the repo's .Brewfile.
func Sync(repoBrewfile string, exec Executor) error {
	tmp, err := os.CreateTemp("", ".Brewfile.current.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpName) }()

	_, err = exec.Run("brew", "bundle", "dump", "--force", "--file="+tmpName)
	if err != nil {
		return fmt.Errorf("brew dump: %w", err)
	}

	out, err := exec.Run("diff", repoBrewfile, tmpName)
	if err != nil {
		var exitErr *osExec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return fmt.Errorf("diff: %w", err)
		}
		// exit code 1 means files differ; fall through to display output
	}
	if out != "" {
		fmt.Print(out)
	} else {
		fmt.Println("no drift — Brewfile matches machine state")
	}
	return nil
}

// Dump writes the live machine brew state back to the repo Brewfile.
func Dump(repoBrewfile string, exec Executor) error {
	_, err := exec.Run("brew", "bundle", "dump", "--force", "--file="+repoBrewfile)
	if err != nil {
		return fmt.Errorf("brew bundle dump: %w", err)
	}
	fmt.Printf("Brewfile updated: %s\n", repoBrewfile)
	return nil
}

// Install installs packages from the repo Brewfile.
func Install(exec Executor) error {
	_, err := exec.Run("brew", "update")
	if err != nil {
		return fmt.Errorf("brew update: %w", err)
	}
	out, err := exec.Run("brew", "bundle", "install", "--global")
	if err != nil {
		if out != "" {
			return fmt.Errorf("brew bundle install: %w\n%s", err, out)
		}
		return fmt.Errorf("brew bundle install: %w", err)
	}
	if out != "" {
		fmt.Print(strings.TrimRight(out, "\n"))
		fmt.Println()
	}
	return nil
}
