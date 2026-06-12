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
