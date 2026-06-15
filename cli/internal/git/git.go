package git

import (
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/dlstadther/bootstrap/cli/internal/version"
)

// EnvGetter looks up an environment variable by key. os.Getenv satisfies it;
// tests pass a fake to exercise RepoPath without touching the real environment.
type EnvGetter func(key string) string

// RepoPath resolves the bootstrap repo path: $BOOTSTRAP_REPO (via getenv) first,
// compiled-in RepoPath as fallback.
func RepoPath(getenv EnvGetter) string {
	if v := getenv("BOOTSTRAP_REPO"); v != "" {
		return v
	}
	return version.RepoPath
}

// CurrentHash returns the HEAD commit hash of the repo at repoPath.
func CurrentHash(repoPath string, exec iexec.Executor) (string, error) {
	out, err := exec.Run("git", "-C", repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsDirty reports whether the repo at repoPath has uncommitted changes.
func IsDirty(repoPath string, exec iexec.Executor) (bool, error) {
	out, err := exec.Run("git", "-C", repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}
