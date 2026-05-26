package brew

import (
	"fmt"
	"strings"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// Sync shows drift between the live machine state and the repo Brewfile.
// repoBrewfile is the absolute path to the repo's .Brewfile.
func Sync(repoBrewfile string, exec Executor) error {
	_, err := exec.Run("brew", "bundle", "dump", "--force", "--file=/tmp/.Brewfile.current")
	if err != nil {
		return fmt.Errorf("brew dump: %w", err)
	}
	out, _ := exec.Run("diff", repoBrewfile, "/tmp/.Brewfile.current")
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
	out, err := exec.Run("brew", "bundle", "install", "--global", "--verbose")
	if err != nil {
		return fmt.Errorf("brew bundle install: %w", err)
	}
	if out != "" {
		fmt.Print(strings.TrimRight(out, "\n"))
		fmt.Println()
	}
	return nil
}
