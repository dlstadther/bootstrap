package sync

import "fmt"

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
